package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	server := InitWebServer()
	server.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello, welcome to here")
	})
	server.Run(":8080")
}
