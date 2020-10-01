package auth

import (
	"context"
	// "encoding/hex"
	"errors"
	"fmt"
	"github.com/lianmi/servers/lmSdkClient/common"
	"github.com/lianmi/servers/util"
	// "github.com/lianmi/servers/internal/pkg/models"
	JSON "github.com/bitly/go-simplejson"
	"net/url"
)

type AuthService struct {
	c *Client
}

//Send request
func (s *AuthService) Do(ctx context.Context, method, endpoint, body string) (*JSON.Json, error) {
	r := &request{
		method:   method,
		endpoint: endpoint,
		query:    url.Values{},
	}

	data, err := s.c.callApi(ctx, r, body)
	if err != nil {
		return nil, err
	}

	j, err := util.NewJSON(data)
	if err != nil {
		return nil, err
	}

	return j, nil
}

// 传入手机号，获取验证码
func (s *AuthService) SendSms(mobile string) (*JSON.Json, error) {
	if mobile == "" {
		return nil, errors.New("Error: mobile is empty.")
	}

	endpoint := fmt.Sprintf(common.ENDPOINT_SMSCODE, mobile)
	responseData, err := s.Do(NewContext(), "GET", endpoint, "")
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

// 传入用户名及密码，短信验证码，登录
func (s *AuthService) Login(login *Login) (*JSON.Json, error) {
	if login.Username == "" {
		return nil, errors.New("Error: Username is empty.")
	}
	if login.Password == "" {
		return nil, errors.New("Error: Password is empty.")
	}
	if login.SmsCode == "" {
		return nil, errors.New("Error: SmsCode is empty.")
	}

	bodyString, err := login.ToJson()

	endpoint := fmt.Sprintf(common.ENDPOINT_LOGIN)
	responseData, err := s.Do(NewContext(), "POST", endpoint, bodyString)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}
