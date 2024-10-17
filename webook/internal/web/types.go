package web

import (
	"github.com/gin-gonic/gin"
	"webook/pkg/ginx"
)

type handler interface {
	RegisterRoutes(server *gin.Engine)
}

// Result 重构的小技巧
type Result = ginx.Result

type Page struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}
