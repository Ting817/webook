package service

import (
	"context"
	"fmt"
	"math/rand"

	"webook/internal/repository"
	"webook/internal/service/sms"
)

const codeTplId = "1877556"

type CodeService struct {
	sms  sms.Service
	repo *repository.CodeRepository
}

func NewCodeService(sms sms.Service, repo *repository.CodeRepository) *CodeService {
	return &CodeService{
		sms:  sms,
		repo: repo,
	}
}

func (cs *CodeService) Send(c context.Context, biz string, phone string) error {
	// 生成一个验证码，然后保持到redis中去，最后发送出去
	code := cs.generateCode()
	err := cs.repo.Store(c, biz, phone, code)
	if err != nil {
		return fmt.Errorf("error store code. %w\n", err)
	}
	err = cs.sms.Send(c, codeTplId, []string{code}, phone)
	if err != nil {
		// redis有验证码，但没发送成功。不能删掉此验证码，因为err有可能是超时问题。
		// 可以重试，在初始化时传入重试的smsSvc
		return fmt.Errorf("error send code. %w\n", err)
	}
	return nil
}

func (cs *CodeService) Verify(c context.Context, biz string, phone string, inputCode string) (bool, error) {
	ok, err := cs.repo.Verify(c, biz, phone, inputCode)
	if err != nil {
		return false, fmt.Errorf("code verify error. %w\n", err)
	}

	return ok, nil
}

func (cs *CodeService) generateCode() string {
	// 随机生成六位数
	num := rand.Intn(999999)
	return fmt.Sprintf("%06d", num)
}
