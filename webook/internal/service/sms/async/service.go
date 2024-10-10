package async

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

//func (s *SMSService) needAsync() bool {
//	// 这边是你要设计的，各种判定要不要触发异步的方案
//	// 1. 基于响应时间的，平均响应时间
//	// 1.1 使用绝对阈值，比如说直接发送的时候，（连续一段时间，或者连续N个请求）响应时间超了
//	// 1.2 变化趋势，比如说当前一秒钟内的所有请求的响应时间比上一秒增长了 X%，就转异步
//	// 2. 基于错误率：一段时间内，收到 err 的请求比率大于 X%，转异步
//
//	// 什么时候退出异步？
//	// 1. 进入异步 N 分钟后
//	// 2. 保留 1% 的流量（或者更少），继续同步发送，判定响应时间/错误率
//	return true
//}
