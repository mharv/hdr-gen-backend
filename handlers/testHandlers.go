package handlers

import (
	"net/http"
	"os/exec"

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
