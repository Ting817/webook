package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	server := InitWebServer()
	server.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello, welcome to here")
	})
	server.Run(":8080")
}
