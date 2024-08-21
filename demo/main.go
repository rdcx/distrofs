package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.LoadHTMLFiles("index.html")

	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.GET("/app1/example.jpg", func(c *gin.Context) {
		c.File("example.jpg")
	})

	r.GET("/app1/script.js", func(c *gin.Context) {
		c.File("script.js")
	})

	r.Run(":8080")
}
