package async

//import (
//	"context"
//	"webook/internal/repository"
//	"webook/internal/service/sms"
//)
//
//type SMSService struct {
//	svc  sms.Service
//	repo repository.SMSAysncReqRepository
//}
//
//func NewSMSService() *SMSService {
//	return &SMSService{}
//}
//
//func (s *SMSService) StartAsync() {
//	go func() {
//		reqs := s.svc.查找没发出去的请求()
//		for _, req := range reqs {
//			// 在这发送 并且控制重试
//			s.svc.Send(, req.biz, req.args, req.numbers...)
//		}
//	}()
//}
//
//func (s *SMSService) Send(c context.Context, biz string, args []string, numbers ...string) error {
//	// 首先是正常路径
//	err := s.svc.Send(c, biz, args, numbers...)
//	if err != nil {
//		// 判定是不是崩溃了,需判断错误率什么的
//
//		if 崩溃了 {
//			s.repo.Store()
//		}
//	}
//	return nil
//}
