package startup

import (
	_ "go.uber.org/zap"
	"webook/pkg/logger"
)

func InitLog() logger.LoggerV1 {
	return logger.NewNoOpLogger()
}
