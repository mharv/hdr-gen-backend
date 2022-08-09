package handlers

import (
	"net/http"
	"os/exec"
	"time"

	"hdr-gen-backend/database"
	"hdr-gen-backend/models"

	"github.com/gin-gonic/gin"
)

func Sleep(c *gin.Context) {
	out, err := exec.Command("./scripts/sleep.sh").Output()

    c.JSON(http.StatusOK, "test")
    return

	response := string(out)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, response)
	}

}

func Rtrace(c *gin.Context) {
	out, err := exec.Command("rtrace", "-defaults").Output()

	response := string(out)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, response)
	}

}

func TestLog(c *gin.Context) {
	var applog models.Applog
	// record metadata in sql
	applog.ProjectId = 0
	applog.ImageId = 0
	applog.Time = time.Now().Format(time.RFC3339)
	applog.Message =  "this is a test, no need to worry"

	if result := database.DB.Create(&applog); result.Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": result.Error,
			"step":    "problem with db write",
		})
		return
	} else {
		c.JSON(http.StatusOK, applog)
    }
}
