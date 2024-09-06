package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"webook/internal/domain"
)

var redirectURI = url.PathEscape("https://meoying.com/oauth2/wechat/callback")

type Service interface {
	AuthURL(c context.Context, state string) (string, error)
	VerifyCode(c context.Context, code string) (domain.WechatInfo, error)
}

type service struct {
	appId     string
	appSecret string
	client    *http.Client
}

//// NewServiceV1 最好的写法
//func NewServiceV1(appId string, appSecret string, client *http.Client) Service {
//	return &service{
//		appId:     appId,
//		appSecret: appSecret,
//		client:    client,
//	}
//}

func NewService(appId string, appSecret string) Service {
	return &service{
		appId:     appId,
		appSecret: appSecret,
		client:    http.DefaultClient, // 未完全依赖注入
	}
}

func (s *service) AuthURL(c context.Context, state string) (string, error) {
	const urlPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=SCOPE&state=%s#wechat_redirect"
	return fmt.Sprintf(urlPattern, s.appId, redirectURI, state), nil
}

func (s *service) VerifyCode(c context.Context, code string) (domain.WechatInfo, error) {
	const targetPattern = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	target := fmt.Sprintf(targetPattern, s.appId, s.appSecret, code)
	//req, err := http.Get(target)
	//req, err := http.NewRequest(http.MethodGet, target, nil)
	req, err := http.NewRequestWithContext(c, http.MethodGet, target, nil)
	if err != nil {
		return domain.WechatInfo{}, fmt.Errorf("Cannot get targetURL. %w\n" + err.Error())
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, fmt.Errorf("error. %w\n" + err.Error())
	}
	decoder := json.NewDecoder(resp.Body)
	var res Result
	err = decoder.Decode(&res)
	if err != nil {
		return domain.WechatInfo{}, fmt.Errorf("decode error. %w\n" + err.Error())
	}
	if res.ErrCode != 0 {
		return domain.WechatInfo{}, fmt.Errorf("errer code status: %d, error msg: %s", res.ErrCode, res.ErrMsg)
	}
	return domain.WechatInfo{
		OpenID:  res.OpenID,
		UnionID: res.UnionID,
	}, nil
}

type Result struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`

	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`

	Scope string `json:"scope"`

	OpenID  string `json:"openid"`  // 在此应用下的唯一ID
	UnionID string `json:"unionid"` // 在此公司下的唯一ID
}
