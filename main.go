package main

import (
	"log"
	"os/exec"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/pong", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ping",
		})
	})
	r.GET("/rtrace", func(c *gin.Context) {
		out, err := exec.Command("date").Output()
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(200, gin.H{
			"message": string(out),
		})
	})
	r.Run(":8080") // listen and serve on 0.0.0.0:8080
}
