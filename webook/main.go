package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"webook/internal/integration"
)

func main() {
	server := integration.InitWebServer()
	server.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello, welcome to here")
	})
	server.Run(":8080")
}
