package tencent

import (
	"context"
	"fmt"
	"go.uber.org/zap"

	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	appId     *string // 取指针 是因为腾讯云SMS的设计要求
	signature *string
	client    *sms.Client
}

func NewService(client *sms.Client, appId string, signature string) *Service {
	return &Service{
		client:    client,
		appId:     ekit.ToPtr[string](appId),
		signature: ekit.ToPtr[string](signature),
	}
}

// Send biz 代表的是tplId
func (s *Service) Send(c context.Context, biz string, args []string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signature
	req.TemplateId = ekit.ToPtr[string](biz)
	req.PhoneNumberSet = s.toStringPtrSlice(numbers)
	req.TemplateParamSet = s.toStringPtrSlice(args)
	resp, err := s.client.SendSms(req)
	zap.L().Debug("调用腾讯短信服务",
		zap.Any("req", req),
		zap.Any("resp", resp))
	if err != nil {
		return fmt.Errorf("error.%w\n", err)
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) == "Ok" {
			return fmt.Errorf("send message failed, code:%s, message:%s", *status.Code, *status.Message)
		}
	}
	return nil
}

func (s *Service) toStringPtrSlice(src []string) []*string {
	return slice.Map[string, *string](src, func(idx int, src string) *string {
		return &src
	})
}
