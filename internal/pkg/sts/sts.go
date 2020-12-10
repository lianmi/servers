package sts

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/satori/go.uuid"
)

type StatementBase struct {
	Effect   string
	Action   []string
	Resource []string
}

type Policy struct {
	Version   string
	Statement []StatementBase
}

func (p *Policy) ToJson() string {
	jsonBytes, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

type AliyunStsClient struct {
	ChildAccountKeyId  string
	ChildAccountSecret string
	RoleAcs            string
}

func NewStsClient(key, secret, roleAcs string) *AliyunStsClient {
	return &AliyunStsClient{
		ChildAccountKeyId:  key,
		ChildAccountSecret: secret,
		RoleAcs:            roleAcs,
	}
}

/*
sessionName - 是一个用来标示临时凭证的名称，一般来说建议使用不同的应用程序用户来区分。
DurationSeconds - 指的是临时凭证的有效期，单位是s，最小为900，最大为3600。
policy -  表示的是在扮演角色的时候额外加上的一个权限限制。此参数可以限制生成的STS token的权限，若不指定则返回的token拥有指定角色的所有权限。
*/
func (cli *AliyunStsClient) GenerateSignatureUrl(sessionName, durationSeconds, policy string) (string, error) {
	assumeUrl := "SignatureVersion=1.0"
	assumeUrl += "&Format=JSON"
	assumeUrl += "&Timestamp=" + url.QueryEscape(time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	assumeUrl += "&RoleArn=" + url.QueryEscape(cli.RoleAcs)
	assumeUrl += "&RoleSessionName=" + sessionName
	assumeUrl += "&AccessKeyId=" + cli.ChildAccountKeyId
	assumeUrl += "&SignatureMethod=HMAC-SHA1"
	assumeUrl += "&Version=2015-04-01"
	assumeUrl += "&Action=AssumeRole"
	assumeUrl += "&SignatureNonce=" + uuid.NewV4().String()
	assumeUrl += "&DurationSeconds=" + durationSeconds

	if policy != "" {
		//TODO Policy 策略，可以精确控制用户的目录权限
		assumeUrl += "&Policy=" + policy //字符串  若policy为空，则用户将获得该角色下所有权限
	}

	// 解析成V type
	signToString, err := url.ParseQuery(assumeUrl)
	if err != nil {
		return "", err
	}

	// URL顺序化
	result := signToString.Encode()

	// 拼接
	StringToSign := "GET" + "&" + "%2F" + "&" + url.QueryEscape(result)

	// HMAC
	hashSign := hmac.New(sha1.New, []byte(cli.ChildAccountSecret+"&"))
	hashSign.Write([]byte(StringToSign))

	// 生成signature
	signature := base64.StdEncoding.EncodeToString(hashSign.Sum(nil))

	// Url 添加signature
	assumeUrl = "https://sts.aliyuncs.com/?" + assumeUrl + "&Signature=" + url.QueryEscape(signature)

	return assumeUrl, nil
}

// 请求构造好的URL,获得授权信息
// TODO: 安全认证 HTTPS
func (cli *AliyunStsClient) GetStsResponse(url string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	return body, err
}
