package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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

	if result := database.DB.Create(&project); result.Error != nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, project)
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
	fullPath := tmpDirName + imageName + "/"

	fmt.Println(imageName)
	os.MkdirAll(fullPath, os.ModePerm)

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
	blobFileName := storage.UploadFileToBlobStore(imageName+".hdr", "/tmp/hdrgen/"+imageName+"/pic/")

	// TODO upload can response curve and exif files
	// TODO upload can response curve and exif files
	// TODO upload can response curve and exif files
	// TODO upload can response curve and exif files

	// save to sql db
	var image models.Image

	image.ProjectId = int32(projectIdInt)
	image.Name = blobFileName
	image.Type = "HDR"

	if result := database.DB.Create(&image); result.Error != nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, image)
	}

}

func goDotEnvVariable(key string) string {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
