package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

	// TODO create hdr file
	out, err := exec.Command("./scripts/runhdr.sh", imageName).Output()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err,
		})
	}

	fmt.Println(string(out))

	// TODO store to blob
	blobFileName := storage.UploadFileToBlobStore(imageName+".hdr", "/tmp/hdrgen/"+imageName+"/pic/")

	// save to sql db
	var image models.Image

	image.ProjectId = 1
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
