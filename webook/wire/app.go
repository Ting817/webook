package wire

import (
	"github.com/gin-gonic/gin"
	"webook/internal/events"
)

type App struct {
	Web       *gin.Engine
	Consumers []events.Consumer
}
