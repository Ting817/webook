package memory

import (
	"context"
	"fmt"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

// Send 模拟发短信的过程 为了测试
func (s Service) Send(c context.Context, tplId string, args []string, numbers ...string) error {
	fmt.Println(args)
	return nil
}
