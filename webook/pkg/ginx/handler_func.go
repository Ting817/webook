package ginx

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func WrapReq[T any](fn func(c *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T
		if err := c.Bind(&req); err != nil {
			c.JSON(http.StatusBadRequest, Result{
				Code: http.StatusBadRequest,
				Msg:  "Invalid request data",
				Data: nil,
			})
			return
		}
		res, err := fn(c, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Result{
				Code: http.StatusInternalServerError,
				Msg:  err.Error(),
				Data: nil,
			})
			return
		}
		c.JSON(http.StatusOK, res)
	}
}

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
