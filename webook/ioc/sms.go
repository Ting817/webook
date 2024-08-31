package ioc

import (
	"webook/internal/service/sms"
	"webook/internal/service/sms/memory"
)

func InitSmsService() sms.Service {
	return memory.NewService()
}
