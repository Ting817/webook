package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/wire"
)

func main() {
	server := wire.InitWebServer()
	server.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello, welcome to here")
	})
	server.Run(":8080")
}
