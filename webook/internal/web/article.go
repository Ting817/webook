package web

import "github.com/gin-gonic/gin"

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
}

func (a ArticleHandler) RegisterRoutes(server *gin.Engine) {

}
