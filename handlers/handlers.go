package handlers

import (
	"net/http"

	"hdr-gen-backend/database"
	"hdr-gen-backend/models"

	"github.com/gin-gonic/gin"
)

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
