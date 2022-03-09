package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"encoding/base64"

	"hdr-gen-backend/database"
	"hdr-gen-backend/models"
	"hdr-gen-backend/storage"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var tmpDirName = goDotEnvVariable("LOCAL_TEMP_UPLOAD_DIRECTORY_NAME")

func GetProjects(c *gin.Context) {
	var projects []models.Project

	if result := database.DB.Find(&projects); result.Error != nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, projects)
	}
}

func PostProject(c *gin.Context) {
	var project models.Project
	c.BindJSON(&project)

	var exists bool
	err := database.DB.Model(project).
		Select("count(*) > 0").
		Where("Number = ?", project.Number).
		Find(&exists).
		Error

	if err != nil {
		fmt.Println(err.Error())
	}

	if !exists {
		fmt.Println("project record does not exist... saving now.")
		if result := database.DB.Create(&project); result.Error != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": result.Error.Error(),
				"step":    "problem with db write",
			})
		} else {
			c.JSON(http.StatusOK, project)
		}
	} else {
		fmt.Println("project record exists... doing nothing.")
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Project with project number %s already exist", project.Number),
		})
	}
}

func GetImages(c *gin.Context) {
	var images []models.Image

	if result := database.DB.Find(&images); result.Error != nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, images)
	}
}

func GetImagesProjectId(c *gin.Context) {
	id := c.Params.ByName("projectId")
	var images []models.Image

	if result := database.DB.Where("ProjectId = ?", id).Find(&images); result.Error != nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, images)
	}
}

func UploadImagesToServer(c *gin.Context) {
	// upload bracketed set, create hdr, store to blob
	projectId := c.Params.ByName("projectId")
	projectIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid projectId, must be an integer",
		})
	}
	imageName := c.Params.ByName("imageName")

	fullPath := createLocalWorkingDirectory(imageName)

	//upload bracketed set
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, "get form err: %s", err.Error())
		return
	}

	files := form.File["files"]
	// The file cannot be received.
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "No file is received",
		})
	}

	for _, file := range files {
		extension := filepath.Ext(file.Filename)
		filename := strings.TrimSuffix(file.Filename, extension)

		newFileName := filename + "-" + uuid.New().String() + extension

		if err := c.SaveUploadedFile(file, fullPath+newFileName); err != nil {
			c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
			return
		}
	}

	// create hdr file
	out, err := exec.Command("./scripts/runhdr.sh", imageName).Output()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err,
		})
	}

	fmt.Println(string(out))

	// store to blob
	blobFileName := storage.UploadFileToBlobStore(imageName+".hdr", fullPath+"pic/", true)

	// TODO upload can response curve and exif files
	// TODO upload can response curve and exif files
	// TODO upload can response curve and exif files
	// TODO upload can response curve and exif files

	// save to sql db
	var image models.Image

	image.ProjectId = int32(projectIdInt)
	image.Name = blobFileName
	image.Type = "HDR"
	image.Status = "ACTIVE"

	if result := database.DB.Create(&image); result.Error != nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, image)
	}

}

func UpExposeImage(c *gin.Context) {
	// upload bracketed set, create hdr, store to blob
	projectId := c.Params.ByName("projectId")
	projectIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid projectId, must be an integer",
		})
		return
	}
	// full image name including extension
	imageName := c.Params.ByName("imageName")
	exposureFactor := c.Params.ByName("exposureFactor")

	extension := filepath.Ext(imageName)
	// full image name without extension
	imageNameOnly := strings.TrimSuffix(imageName, extension)

	fullPath := createLocalWorkingDirectory(imageName)

	fmt.Printf("image name: %s \nprojectId: %d \nexposure factor: %s", imageName, projectIdInt, exposureFactor)
	// load current HDR to tmp dir
	storage.DownloadFileToLocalDir(imageName, fullPath+"pic/")

	// run upexpose script
	out, err := exec.Command("./scripts/upexpose.sh", imageNameOnly, exposureFactor).Output()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"state":   "at execute script",
		})
		return
	}

	fmt.Println(string(out))

	// upload result to storage with ACTIVE status
	blobFileName := storage.UploadFileToBlobStore(imageName, fullPath+"pic/", false)

	jpgblobFileName := storage.UploadFileToBlobStore(imageNameOnly+"-exposed.jpg", fullPath+"tif/", false)

	// record metadata in sql
	var image models.Image

	image.ProjectId = int32(projectIdInt)
	image.Name = jpgblobFileName
	image.Type = "EXPOSED"
	image.Status = "ACTIVE"

	var exists bool
	err = database.DB.Model(image).
		Select("count(*) > 0").
		Where("Name = ?", image.Name).
		Find(&exists).
		Error
	if !exists {
		fmt.Println("image record does not exist... saving now.")
		if result := database.DB.Create(&image); result.Error != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": err.Error(),
				"step":    "problem with db write",
			})
			return
		}
	} else {
		fmt.Println("image record exists... doing nothing.")
	}

	// get the exposed photo as a base64 encoded jpg and return in request
	data, err := ioutil.ReadFile(fullPath + "tif/" + imageNameOnly + "-exposed.jpg")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err,
		})
	}

	var base64Encoding string

	base64Encoding += "data:image/jpeg;base64,"
	base64Encoding += base64.StdEncoding.EncodeToString(data)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("image %s has been uploaded to storage.", blobFileName),
		"image":   base64Encoding,
	})
}

func DownExposeImage(c *gin.Context) {
	// upload bracketed set, create hdr, store to blob
	projectId := c.Params.ByName("projectId")
	projectIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid projectId, must be an integer",
		})
	}
	// full image name including extension
	imageName := c.Params.ByName("imageName")
	exposureFactor := c.Params.ByName("exposureFactor")

	extension := filepath.Ext(imageName)
	// full image name without extension
	imageNameOnly := strings.TrimSuffix(imageName, extension)

	fullPath := createLocalWorkingDirectory(imageName)

	fmt.Printf("image name: %s \nprojectId: %d \nexposure factor: %s", imageName, projectIdInt, exposureFactor)
	// load current HDR to tmp dir
	storage.DownloadFileToLocalDir(imageName, fullPath+"pic/")

	// run upexpose script
	out, err := exec.Command("./scripts/downexpose.sh", imageNameOnly, exposureFactor).Output()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err,
		})
	}

	fmt.Println(string(out))

	// upload result to storage with ACTIVE status
	blobFileName := storage.UploadFileToBlobStore(imageName, fullPath+"pic/", false)

	jpgblobFileName := storage.UploadFileToBlobStore(imageNameOnly+"-exposed.jpg", fullPath+"tif/", false)

	// record metadata in sql
	var image models.Image

	image.ProjectId = int32(projectIdInt)
	image.Name = jpgblobFileName
	image.Type = "EXPOSED"
	image.Status = "ACTIVE"

	var exists bool
	err = database.DB.Model(image).
		Select("count(*) > 0").
		Where("Name = ?", image.Name).
		Find(&exists).
		Error
	if !exists {
		fmt.Println("image record does not exist... saving now.")
		if result := database.DB.Create(&image); result.Error != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": err.Error(),
				"step":    "problem with db write",
			})
			return
		}
	} else {
		fmt.Println("image record exists... doing nothing.")
	}

	// get the exposed photo as a base64 encoded jpg and return in request
	data, err := ioutil.ReadFile(fullPath + "tif/" + imageNameOnly + "-exposed.jpg")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err,
		})
	}

	var base64Encoding string

	base64Encoding += "data:image/jpeg;base64,"
	base64Encoding += base64.StdEncoding.EncodeToString(data)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("image %s has been uploaded to storage.", blobFileName),
		"image":   base64Encoding,
	})
}

func LuminanceMatrix(c *gin.Context) {
	// upload bracketed set, create hdr, store to blob
	projectId := c.Params.ByName("projectId")
	projectIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid projectId, must be an integer",
		})
	}
	// full image name including extension
	imageName := c.Params.ByName("imageName")
	extension := filepath.Ext(imageName)
	// full image name without extension
	imageNameOnly := strings.TrimSuffix(imageName, extension)

	fullPath := createLocalWorkingDirectory(imageName)

	// load current HDR to tmp dir
	storage.DownloadFileToLocalDir(imageName, fullPath+"pic/")

	// run matrix script
	out, err := exec.Command("./scripts/matrix.sh", imageNameOnly).Output()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"step":    "matrix.sh",
		})
		return
	}

	fmt.Println(string(out))

	// upload result to storage
	// deactivated for testing
	blobFileName := storage.UploadFileToBlobStore(imageName, fullPath+"pic/", false)

	jpgblobFileName := storage.UploadFileToBlobStore(imageNameOnly+"-scaled.jpg", fullPath+"tif/", false)

	// record metadata in sql
	var image models.Image

	image.ProjectId = int32(projectIdInt)
	image.Name = jpgblobFileName
	image.Type = "SCALED"
	image.Status = "ACTIVE"

	if result := database.DB.Create(&image); result.Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
			"step":    "problem with db write",
		})
		return
	}

	// get the exposed photo as a base64 encoded jpg and return in request
	data, err := ioutil.ReadFile(fullPath + "tif/" + imageNameOnly + "-scaled.jpg")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"step":    "reading generated jpg",
		})
		return
	}

	var base64Encoding string

	base64Encoding += "data:image/jpeg;base64,"
	base64Encoding += base64.StdEncoding.EncodeToString(data)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("image %s has been uploaded to storage.", blobFileName),
		"image":   base64Encoding,
	})
}

func ScaleImage(c *gin.Context) {
	// upload bracketed set, create hdr, store to blob
	// projectId := c.Params.ByName("projectId")
	// projectIdInt, err := strconv.Atoi(projectId)
	// if err != nil {
	// 	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
	// 		"message": "Invalid projectId, must be an integer",
	// 	})
	// 	return
	// }
	currentLuminanceLevel := c.Params.ByName("current")
	targetLuminanceLevel := c.Params.ByName("target")
	currentLuminanceLevelFloat, err := strconv.ParseFloat(currentLuminanceLevel, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid currentLuminanceLevel, must be a float",
		})
		return
	}
	targetLuminanceLevelFloat, err := strconv.ParseFloat(targetLuminanceLevel, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid currentLuminanceLevel, must be a float",
		})
		return
	}

	// calculate scale factor based of current and target readings
	scaleFactor := fmt.Sprintf("%f", ((targetLuminanceLevelFloat / currentLuminanceLevelFloat) * 1))

	// full image name including extension
	imageName := c.Params.ByName("imageName")
	extension := filepath.Ext(imageName)
	// full image name without extension
	imageNameOnly := strings.TrimSuffix(imageName, extension)

	fullPath := createLocalWorkingDirectory(imageName)

	// load current HDR to tmp dir
	storage.DownloadFileToLocalDir(imageName, fullPath+"pic/")

	fmt.Printf("imageNameOnly, %s \n scaleFactor, %s \n fullPath %s \n", imageNameOnly, scaleFactor, fullPath)

	// run matrix script
	out, err := exec.Command("./scripts/scaling.sh", imageNameOnly, scaleFactor, fullPath).Output()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"step":    "scaling.sh",
		})
		return
	}

	fmt.Println(string(out))

	// upload result to storage
	// deactivated for testing
	blobFileName := storage.UploadFileToBlobStore(imageName, fullPath+"pic/", false)

	jpgblobFileName := storage.UploadFileToBlobStore(imageNameOnly+"-scaled.jpg", fullPath+"tif/", false)
	fmt.Println(jpgblobFileName)

	// get the exposed photo as a base64 encoded jpg and return in request
	data, err := ioutil.ReadFile(fullPath + "tif/" + imageNameOnly + "-scaled.jpg")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"step":    "error reading generated jpg",
		})
		return
	}

	// TODO do we record scling factor to mysql here?

	var base64Encoding string

	base64Encoding += "data:image/jpeg;base64,"
	base64Encoding += base64.StdEncoding.EncodeToString(data)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("image %s has been uploaded to storage.", blobFileName),
		"image":   base64Encoding,
	})
}

func FalseColour(c *gin.Context) {
	// upload bracketed set, create hdr, store to blob
	projectId := c.Params.ByName("projectId")
	projectIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid projectId, must be an integer",
		})
	}
	// full image name including extension
	imageName := c.Params.ByName("imageName")
	extension := filepath.Ext(imageName)
	// full image name without extension
	imageNameOnly := strings.TrimSuffix(imageName, extension)

	fullPath := createLocalWorkingDirectory(imageName)

	// load current HDR to tmp dir
	storage.DownloadFileToLocalDir(imageName, fullPath+"pic/")

	// run matrix script
	out, err := exec.Command("./scripts/falsecolour.sh", imageNameOnly).Output()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"step":    "falsecolour.sh",
		})
		return
	}

	fmt.Println(string(out))

	// upload result to storage
	// deactivated for testing
	blobFileName := storage.UploadFileToBlobStore(imageNameOnly+"-falseColor.jpg", fullPath+"tif/", false)

	// record metadata in sql
	var image models.Image

	image.ProjectId = int32(projectIdInt)
	image.Name = blobFileName
	image.Type = "FALSECOLOR"
	image.Status = "ACTIVE"

	if result := database.DB.Create(&image); result.Error != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "false color generation complete.",
	})
}

func createLocalWorkingDirectory(imageName string) string {
	extension := filepath.Ext(imageName)
	// full image name without extension
	imageNameOnly := strings.TrimSuffix(imageName, extension)

	fullPath := tmpDirName + imageNameOnly + "/"
	os.MkdirAll(fullPath+"pic/", os.ModePerm)
	os.MkdirAll(fullPath+"tmp/", os.ModePerm)
	os.MkdirAll(fullPath+"tif/", os.ModePerm)
	os.MkdirAll(fullPath+"exif/", os.ModePerm)
	return fullPath
}

func goDotEnvVariable(key string) string {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
