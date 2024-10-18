package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/wire"
)

func main() {
	app := wire.InitApp()
	for _, c := range app.Consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	server := app.Web
	server.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello, welcome to here")
	})
	server.Run(":8080")
}
