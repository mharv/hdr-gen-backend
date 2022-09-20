package handlers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"encoding/base64"

	"hdr-gen-backend/database"
	"hdr-gen-backend/models"
	"hdr-gen-backend/storage"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var tmpDirName = goDotEnvVariable("LOCAL_TEMP_UPLOAD_DIRECTORY_NAME")

func GetProjectByNumber(c *gin.Context) {
	projectNumber := c.Params.ByName("projectNumber")
	var project models.Project

	if result := database.DB.Where("Number = ?", projectNumber).Find(&project); result.Error != nil {
        logMessage(-1, -1, fmt.Sprintf("Project with project number %s cannot be found.", projectNumber))
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, project)
	}
}

func GetProjects(c *gin.Context) {
	var projects []models.Project

	if result := database.DB.Find(&projects); result.Error != nil {
        logMessage(-1, -1, "Projects could not be loaded.")
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
        logMessage(-1, -1, "Error creating project in DB.")
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": result.Error.Error(),
				"step":    "problem with db write",
			})
		} else {
			c.JSON(http.StatusOK, project)
		}
	} else {
		fmt.Println("project record exists... doing nothing.")
        logMessage(-1, -1, fmt.Sprintf("Project with project number %s already exists.", project.Number))
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Project with project number %s already exist", project.Number),
		})
	}
}

func GetImages(c *gin.Context) {
	var images []models.Image

	if result := database.DB.Find(&images); result.Error != nil {
        logMessage(-1, -1, "Error getting image records from DB.")
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, images)
	}
}

func GetImagesProjectId(c *gin.Context) {
	id := c.Params.ByName("projectId")
	var images []models.Image

    id_number, err := strconv.ParseInt(id, 10, 32)
    if err != nil {
        panic(err)
    }
	// get list of images with projectId
	if result := database.DB.Where("ProjectId = ? AND Type = 'HDR'", id).Find(&images); result.Error != nil {
        logMessage(int32(id_number), -1, "Could not get images for project.")
		c.AbortWithStatus(http.StatusNotFound)

		cleanup(tmpDirName)
		return
	}

	var imageNames []string

	for _, image := range images {
		//remove extension from image name before append
		var tempName string
		tempName = strings.Replace(image.Name, ".hdr", "", 1)
		imageNames = append(imageNames, tempName)
	}

	type imageOutput struct {
		Name     string
		Encoding string
	}

	var imageLists [][]imageOutput


	for _, imageName := range imageNames {
		if result := database.DB.Where("Name LIKE ? AND Type <> 'HDR'", imageName+"%").Find(&images); result.Error != nil {
		// if result := database.DB.Where("Name LIKE ? AND Type <> 'HDR' AND Type <> 'BASE'", imageName+"%").Find(&images); result.Error != nil {
            logMessage(int32(id_number), -1, fmt.Sprintf("Could not get images for image with image name %s.", imageName))
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			if len(images) > 0 {
				// create dir with imageName
				// full image name without extension
				fullPath := createLocalWorkingDirectory(imageName)

				var tempImageArray []imageOutput

                // the following code checks if both a base image and an exposed image is in the image list.
                exposedFlag := 0;
                baseIndex := 0;
				for i, image := range images {
                    if strings.Contains(image.Name, "exposed") {
                        exposedFlag = 1
                    }
                    if strings.Contains(image.Name, "base") {
                        baseIndex = i
                    }
                }

                // if both are in the list, remove the base image as the exposed will be used for display.

                if exposedFlag == 1 {
                    images = append(images[:baseIndex], images[baseIndex+1:]...)
                }

				for _, image := range images {
					//download image
					// fmt.Println(fullPath + "tif/" + imageName)
					storage.DownloadFileToLocalDir(image.Name, fullPath+"tif/")

					// get base 64 encoding
					data, err := ioutil.ReadFile(fullPath + "tif/" + image.Name)
					if err != nil {
                        logMessage(int32(id_number), -1, fmt.Sprintf("Could not read generated jpg image with image name %s.", image.Name))
						c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
							"message": err.Error(),
							"step":    "reading generated jpg",
						})

						cleanup(tmpDirName)
						return
					}

					var base64Encoding string

					base64Encoding += "data:image/jpeg;base64,"
					base64Encoding += base64.StdEncoding.EncodeToString(data)

					var tempImage imageOutput

					tempImage.Name = image.Name
					tempImage.Encoding = base64Encoding

					tempImageArray = append(tempImageArray, tempImage)
				}

				imageLists = append(imageLists, tempImageArray)
			}
		}
	}

    fmt.Println("imageLists")
    for _, name := range imageLists {
        for _, item := range name {
            fmt.Println(item.Name)
        }
    }
    fmt.Println("END_")

	cleanup(tmpDirName)
	c.JSON(http.StatusOK, imageLists)
}

func DownloadImagesProjectId(c *gin.Context) {
	id := c.Params.ByName("projectId")
	var images []models.Image


    id_number, err := strconv.ParseInt(id, 10, 32)
    if err != nil {
        panic(err)
    }

	var projectNumber string
	var project models.Project

    if result := database.DB.Where("Id = ?", id).Find(&project); result.Error != nil {
        logMessage(int32(id_number), -1, fmt.Sprintf("project with id %s could not be found in DB.", id))
        c.AbortWithStatus(http.StatusNotFound)

        cleanup(tmpDirName)
        return
    } else {
        projectNumber = project.Number
    }
    // set header for file reponse
	c.Writer.Header().Set("Content-type", "application/octet-stream")

	// add project number in eventually
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-NVA-HDR.zip", projectNumber))
	// create zip writer
	ar := zip.NewWriter(c.Writer)

	// get list of images with projectId
	if result := database.DB.Where("ProjectId = ? AND Type = 'HDR'", id).Find(&images); result.Error != nil {
        logMessage(int32(id_number), -1, fmt.Sprintf("could not get image names from DB for project %s", projectNumber))
		c.AbortWithStatus(http.StatusNotFound)

		cleanup(tmpDirName)
		return
	}

	var imageNames []string

	for _, image := range images {
		//remove extension from image name before append
		var tempName = strings.Replace(image.Name, ".hdr", "", 1)
		imageNames = append(imageNames, tempName)
	}

	for _, imageName := range imageNames {
		if result := database.DB.Where("Name LIKE ? AND Type <> 'HDR' AND Type <> 'BASE'", imageName+"%").Find(&images); result.Error != nil {
        logMessage(int32(id_number), -1, fmt.Sprintf("could not get image from DB for image %s", imageName))
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			if len(images) > 0 {
				// create dir with imageName
				// full image name without extension
				fullPath := createLocalWorkingDirectory(imageName)

				// create zip here
				for _, image := range images {
					//download image
					storage.DownloadFileToLocalDir(image.Name, fullPath+"tif/")

					tempFile, _ := os.Open(fullPath + "tif/" + image.Name)
					tempFileArchived, _ := ar.Create(image.Name)
					io.Copy(tempFileArchived, tempFile)
				}
			}
		}
	}
	ar.Close()
	cleanup(tmpDirName)
	c.JSON(http.StatusOK, "zip of images returned")
}

type LuminanceAverageResponse struct {
    Image string
}

func PostLuminanceAverages(c *gin.Context) {
	projectId := c.Params.ByName("projectId")
	projectIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid projectId, must be an integer",
		})

		cleanup(tmpDirName)
		return
	}

    imageName := c.Params.ByName("imageName")

	fullPath := createLocalWorkingDirectory(imageName)

    var luminanceAverage LuminanceAverageResponse
	c.BindJSON(&luminanceAverage)

    imageNameOnly := strings.Split(imageName, ".")[0]

    idx := strings.Index(luminanceAverage.Image, ";base64,")
    if idx < 0 {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("base64 data of luminance averages image is invalid for project %s", projectId))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid base64 image data",
		})
		cleanup(tmpDirName)
		return
    }

    unbased, err := base64.StdEncoding.DecodeString(luminanceAverage.Image[idx+8:])
    r := bytes.NewReader(unbased)

    im, err := jpeg.Decode(r)
    if err != nil {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("bad jpeg for project %s", projectId))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "bad jpeg image data, base64 decoding failed",
		})
        fmt.Printf("err: %s",err)
		cleanup(tmpDirName)
		return
    }

    fmt.Printf("fullpath:  %s \n", fullPath)
    f, err := os.OpenFile(fullPath+"tif/"+imageNameOnly+"-luminance-averages.jpg", os.O_WRONLY|os.O_CREATE, 0777)
    if err != nil {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("opening luminance average image failed for project %s", projectId))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "opening luminance average image failed , base64 decoding failed",
		})
        fmt.Printf("err: %s",err)
		cleanup(tmpDirName)
		return
    }

    jpeg.Encode(f, im, nil)



    // store image in azure storage
    fmt.Println(fullPath+"tif/"+imageNameOnly+"-luminance-averages.jpg")
	blobFileName := storage.UploadFileToBlobStore(imageNameOnly+"-luminance-averages.jpg", fullPath+"tif/", false)

    fmt.Printf("blobFileName: %s", blobFileName)
    // save image metadata to DB

	var image models.Image

	image.ProjectId = int32(projectIdInt)
	image.Name = blobFileName
	image.Type = "LUMAVG"
	image.Status = "ACTIVE"

	if result := database.DB.Create(&image); result.Error != nil {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("error saving LUMAVG image to DB for project %s", projectId))
		c.AbortWithStatus(http.StatusNotFound)
		cleanup(tmpDirName)
		return
	}

	c.JSON(http.StatusOK, luminanceAverage.Image)
}

func UploadImagesToServer(c *gin.Context) {
	// upload bracketed set, create hdr, store to blob
	projectId := c.Params.ByName("projectId")
	projectIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid projectId, must be an integer",
		})

		cleanup(tmpDirName)
		return
	}
	imageName := c.Params.ByName("imageName")

	fullPath := createLocalWorkingDirectory(imageName)

	//upload bracketed set
	form, err := c.MultipartForm()
	if err != nil {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("get form err: %s", projectId))
		c.String(http.StatusBadRequest, "get form err: %s", err.Error())

		cleanup(tmpDirName)
		return
	}

	files := form.File["files"]

	for i, file := range files {
		extension := filepath.Ext(file.Filename)
		// filename := strings.TrimSuffix(file.Filename, extension)

        filename := fmt.Sprintf("subset%d", i+1)
		newFileName := filename + "-" + uuid.New().String() + extension

		if err := c.SaveUploadedFile(file, fullPath+newFileName); err != nil {
            logMessage(int32(projectIdInt), -1, fmt.Sprintf("error uploading file to tmp filesystem for project %s", projectId))
			c.String(http.StatusBadRequest, "upload file err: %s", err.Error())

			cleanup(tmpDirName)
			return
		}
	}


	var project models.Project

	if result := database.DB.Where("Id = ?", projectId).Find(&project); result.Error != nil {
        logMessage(-1, -1, fmt.Sprintf("Project with project Id: %s cannot be found.", projectId))
		c.AbortWithStatus(http.StatusNotFound)
	}

    fmt.Println("response curve:")
    fmt.Println(project.ResponseCurve)
    fmt.Println("END")

    if project.ResponseCurve == "" {
        // select insta360-melbourne.rsp as default
        project.ResponseCurve = "insta360-melbourne.rsp"
    }


	// copy response curve to /tmp/hdrgen/{name}
	responseCurveFileString := "./responseCurves/" + project.ResponseCurve
	imageTmpDirString := tmpDirName + imageName

	_, err = exec.Command("cp", responseCurveFileString, imageTmpDirString).Output()
	if err != nil {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("error copying response curve for project %s", projectId))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})

		cleanup(tmpDirName)
		return
	}

	// create hdr file
    out, err := exec.Command("./scripts/runhdr.sh", imageName, tmpDirName, project.ResponseCurve).Output()
	if err != nil {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("error creating hdr file for image %s, project %s", imageName, projectId))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
            "stuck" : "here",
			"message": err.Error(),
		})

		cleanup(tmpDirName)
		return
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
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("error saving HDR image to DB for project %s", projectId))
		c.AbortWithStatus(http.StatusNotFound)
		cleanup(tmpDirName)
		return
	}

	extension := filepath.Ext(blobFileName)
	// full image name without extension
	imageNameOnly := strings.TrimSuffix(blobFileName, extension)
	fmt.Println(blobFileName)
	fmt.Println(imageNameOnly)

	// rename .jpg
	Original_Path := fullPath + "tif/" + imageName + "-base.jpg"
	New_Path := fullPath + "tif/" + imageNameOnly + "-base.jpg"
	e := os.Rename(Original_Path, New_Path)
	if e != nil {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("error renaming image path for project %s", projectId))
		log.Fatal(e)
	}

	jpgblobFileName := storage.UploadFileToBlobStore(imageNameOnly+"-base.jpg", fullPath+"tif/", false)

	// record metadata in sql
	image.ProjectId = int32(projectIdInt)
	image.Name = jpgblobFileName
	image.Type = "BASE"
	image.Status = "ACTIVE"

	if result := database.DB.Create(&image); result.Error != nil {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("error saving BASE image to DB for project %s", projectId))
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
			"step":    "problem with db write",
		})

		cleanup(tmpDirName)
		return
	}

	cleanup(tmpDirName)
	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("image %s has been uploaded to storage.", blobFileName),
		"previewImage": jpgblobFileName,
		"hdrImage":     blobFileName,
	})
}

func GetImageByName(c *gin.Context) {
	imageName := c.Params.ByName("imageName")

	extension := filepath.Ext(imageName)
	// full image name without extension
	imageNameOnly := strings.TrimSuffix(imageName, extension)

	fullPath := createLocalWorkingDirectory(imageNameOnly)

	//get image downloaded

	storage.DownloadFileToLocalDir(imageName, fullPath+"tif/")

	// get the exposed photo as a base64 encoded jpg and return in request
	data, err := ioutil.ReadFile(fullPath + "tif/" + imageName)
	if err != nil {
        logMessage(-1, -1, fmt.Sprintf("could not get image %s", imageName))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"step":    "reading generated jpg",
		})

		cleanup(tmpDirName)
		return
	}

	//return base64
	var base64Encoding string

	base64Encoding += "data:image/jpeg;base64,"
	base64Encoding += base64.StdEncoding.EncodeToString(data)

	//cleanup

	cleanup(tmpDirName)
	c.JSON(http.StatusOK, gin.H{
		"image": base64Encoding,
	})
}

func UpExposeImage(c *gin.Context) {
	// upload bracketed set, create hdr, store to blob
	projectId := c.Params.ByName("projectId")
	projectIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid projectId, must be an integer",
		})

		cleanup(tmpDirName)
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
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("error running upexpose script on image %s for project %s",imageName, projectId))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"state":   "at execute script",
		})

		cleanup(tmpDirName)
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

			cleanup(tmpDirName)
			return
		}
	} else {
		fmt.Println("image record exists... doing nothing.")
	}

	// get the exposed photo as a base64 encoded jpg and return in request
	data, err := ioutil.ReadFile(fullPath + "tif/" + imageNameOnly + "-exposed.jpg")
	if err != nil {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("failed to read image %s for project %s",imageNameOnly, projectId))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err,
		})

		cleanup(tmpDirName)
		return
	}

	var base64Encoding string

	base64Encoding += "data:image/jpeg;base64,"
	base64Encoding += base64.StdEncoding.EncodeToString(data)

	cleanup(tmpDirName)
	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("image %s has been uploaded to storage.", blobFileName),
		"image":        base64Encoding,
		"previewImage": jpgblobFileName,
		"hdrImage":     blobFileName,
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

		cleanup(tmpDirName)
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
	out, err := exec.Command("./scripts/downexpose.sh", imageNameOnly, exposureFactor).Output()
	if err != nil {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("error running downexpose script on image %s for project %s",imageName, projectId))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err,
		})

		cleanup(tmpDirName)
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

			cleanup(tmpDirName)
			return
		}
	} else {
		fmt.Println("image record exists... doing nothing.")
	}

	// get the exposed photo as a base64 encoded jpg and return in request
	data, err := ioutil.ReadFile(fullPath + "tif/" + imageNameOnly + "-exposed.jpg")
	if err != nil {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("failed to read image %s for project %s",imageNameOnly, projectId))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err,
		})
		cleanup(tmpDirName)
		return
	}

	var base64Encoding string

	base64Encoding += "data:image/jpeg;base64,"
	base64Encoding += base64.StdEncoding.EncodeToString(data)

	cleanup(tmpDirName)

	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("image %s has been uploaded to storage.", blobFileName),
		"image":        base64Encoding,
		"previewImage": jpgblobFileName,
		"hdrImage":     blobFileName,
	})
}

func LuminanceLevels(c *gin.Context) {
	// upload bracketed set, create hdr, store to blob
	projectId := c.Params.ByName("projectId")
	projectIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid projectId, must be an integer",
		})

		cleanup(tmpDirName)
		return
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
	out, err := exec.Command("./scripts/luminanceLevels.sh", imageNameOnly).Output()
	if err != nil {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("failed to get luminance matrix for image %s for project %s",imageNameOnly, projectId))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"step":    "matrix.sh",
		})

		cleanup(tmpDirName)
		return
	}

	fmt.Println(string(out))

	temp := strings.Split(string(out), "\n")

	type key struct {
		x string
		y string
	}

	response := make(map[string]float64)

    sum := 0.0
    count := 0.0

	for _, record := range temp {
		if len(record) > 0 {
			fragmented := strings.Split(record, " ")

			f1, err := strconv.ParseFloat(fragmented[2], 32)
			if err != nil {
				fmt.Println(err)
			}

			tempKey := fmt.Sprintf("%s, %s", fragmented[0], fragmented[1])
			response[tempKey] = f1

            sum += f1
            count += 1

		}
	}
    response["average"] =  sum / count
    response["average"] = math.Floor(response["average"]*100)/100

	response["x"] = 500
	response["y"] = 350

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("image %d has been uploaded to storage.", projectIdInt),
		"data":    response,
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

		cleanup(tmpDirName)
		return
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
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("failed to get luminance matrix for image %s for project %s",imageNameOnly, projectId))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"step":    "matrix.sh",
		})

		cleanup(tmpDirName)
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
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("failed to save image %s metadata for project %s to DB", image.Name, projectId))
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
			"step":    "problem with db write",
		})

		cleanup(tmpDirName)
		return
	}

	cleanup(tmpDirName)
	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("image %s has been uploaded to storage.", blobFileName),
		"previewImage": jpgblobFileName,
		"hdrImage":     blobFileName,
	})
}

func ScaleImage(c *gin.Context) {
	currentLuminanceLevel := c.Params.ByName("current")
	targetLuminanceLevel := c.Params.ByName("target")
	currentLuminanceLevelFloat, err := strconv.ParseFloat(currentLuminanceLevel, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid currentLuminanceLevel, must be a float",
		})

		cleanup(tmpDirName)
		return
	}

	targetLuminanceLevelFloat, err := strconv.ParseFloat(targetLuminanceLevel, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid currentLuminanceLevel, must be a float",
		})

		cleanup(tmpDirName)
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
	out, err := exec.Command("./scripts/scaleImage.sh", imageNameOnly, scaleFactor, fullPath).Output()
	if err != nil {
        logMessage(-1, -1, fmt.Sprintf("failed to scale image %s .", imageNameOnly))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"step":    "scaleImage.sh",
		})

		cleanup(tmpDirName)
		return
	}

	fmt.Println(string(out))

	// upload result to storage
	// deactivated for testing
	blobFileName := storage.UploadFileToBlobStore(imageName, fullPath+"pic/", false)

	cleanup(tmpDirName)
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("image %s has been uploaded to storage.", blobFileName),
		// "image":   base64Encoding,
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

		cleanup(tmpDirName)
		return
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
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("failed to generate false colour image %s .", imageNameOnly))
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"step":    "falsecolour.sh",
		})

		cleanup(tmpDirName)
		return
	}

	fmt.Println(string(out))

	// upload result to storage
	// deactivated for testing
	blobFileName1 := storage.UploadFileToBlobStore(imageNameOnly+"-falseColor1.jpg", fullPath+"tif/", false)

	// record metadata in sql
	var image models.Image

	image.ProjectId = int32(projectIdInt)
	image.Name = blobFileName1
	image.Type = "FALSECOLOR"
	image.Status = "ACTIVE"

	if result := database.DB.Create(&image); result.Error != nil {
		c.AbortWithStatus(http.StatusNotFound)

		cleanup(tmpDirName)
		return
	}

	blobFileName2 := storage.UploadFileToBlobStore(imageNameOnly+"-falseColor2.jpg", fullPath+"tif/", false)

	image.ProjectId = int32(projectIdInt)
	image.Name = blobFileName2
	image.Type = "FALSECOLOR"
	image.Status = "ACTIVE"

	if result := database.DB.Create(&image); result.Error != nil {
        logMessage(int32(projectIdInt), -1, fmt.Sprintf("failed to save false colour metadata for image %s .", imageNameOnly))
		c.AbortWithStatus(http.StatusNotFound)

		cleanup(tmpDirName)
		return
	}

	cleanup(tmpDirName)

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

func cleanup(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatal(err)
	}
}

func goDotEnvVariable(key string) string {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func printDir(path string) {
    files, err := ioutil.ReadDir(path)
    if err != nil {
        log.Fatal(err)
    }

    for _, file := range files {
        fmt.Println(file.Name(), file.IsDir())
    }
}

func logMessage(projectId, imageId int32, message string) string {
	var applog models.Applog
	// record metadata in sql
    if projectId != -1 {
        applog.ProjectId = projectId
    }
    if imageId != -1 {
        applog.ImageId = imageId
    }
	applog.Time = time.Now().Format(time.RFC3339)
	applog.Message = message

	if result := database.DB.Create(&applog); result.Error != nil {
		return "messaged successfully logged."
	} else {
		return "messaged was not logged..."
    }
}
