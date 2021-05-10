package main

import (
	"encoding/json"
	"fmt"
	"github.com/iGoogle-ink/gopay"
	"github.com/iGoogle-ink/gopay/pkg/util"
	"github.com/iGoogle-ink/gopay/pkg/xlog"
	"github.com/iGoogle-ink/gopay/wechat"
	//"test_wx_pay/wechat"
)

const (
	WechatPay_appID  = "wx9dff85e1c3a3b342"
	WechatPay_apiKey = "LianmiLianmiLianmiLianmicloud508"
	WechatPay_mchId  = "1608460662"
	//WechatPay_mchId            = "1608720021"
	//WechatPay_SUBmchId            = "1608720021"

	WechatPay_SUBmchId         = "1608737479"
	WechatPay_SUBmchId_LM      = "1608720021"
	WechatPay_SUBAppid         = "wx239c33b9be7cd047"
	WechatPay_SUBAppid_LM      = "wx9dff85e1c3a3b342"
	WechatPay_platformCertPath = "./configs/apiclient_cert.p12"

	Wechat_apiV3Key = "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCwHsxgxRWOSAug\nWORUl7LhRlGaDcy0UyplsDVAafWX7jucHHYnaQ0z5bc7oHt2xMxYWirnQoVRKfK0\noHXxclBfM+m3+0Ld/kEUq92AAvk2XyMF1OPP8c8b2D99b7+qYrDswZkSqD1ip8U6\nQVj3yy4MI8mXULCUciel1O/o9ZERg9PRnnC8LeZmHPYQn0xlsS2UyyZrjwpfdwDQ\nzZWp+pDgoM4/JD6ng8+MOeCIgpkrv7j7MlJV5RhsPHaxGB1meqFSFOFzRYj7mvKr\nCLydJPHqW7smxlSmupL2eDkh6ppzGtK9DKlr/gliHx8oxHF4XmD53qL68z96WUTG\nsWsyp8PXAgMBAAECggEAJEgj+GeBek8zPfQyDY82xZvT4bWoDxn26P306nEprAPi\n+dUPLi1BEAjpx3nXFW+TXTwuDHgyuLa4jidkRLo0/nfWVRWI/+yKAbUqK13xcxcE\nQwZJbCQ3c1gINFIaHQK5yfxUCXjpNHK1ebvOlTWhJdUViXuQ9PPTYSFNzyMPoJi9\nQRXr57VEpiPifFeKweFv6oHAX0G05Ng0IOrvhK0zq1Ca9xUohjuOO5EhMXugWFBk\nrmOK834ECEoDeAPujQIDvt6ShoUPPKb55whkcw+wEyhIDId/0Z5ByLb3NdwdW+xX\nMwX5CMB0rytdT9anbheg1xuYc0OHOCm3EYyyIb8dGQKBgQDYLtsQDuCfepNUDaCu\nMb7k+iAxYbH/xNm0GVGq0y1CS4B3kdH5sU4U/P5i34d9+cOW4f8xyXQDSiF9KwhH\nAvXdEtTMMuC5eZWIfTgQPFyGDfXnbJAfhQiHcp2thZuGeBzB6BLg1BPw8V76BlWT\nTUJ5rjvRBUNBDCNv+1U63FfZvQKBgQDQjvbIWMlzEuRGtR5z2o7VhwyERiQH2BYC\nbEzdjetHVFRYRpXWPODKhdbsXXEDMOi3dnel1CesIyW7AcJs9rAA7R8tXm/QGeYx\nimYZqfP0WWK2pNnDarvluacCXsDmbx5klAd0Ka59s8dcO6I17WNVPVSrQwg2fJp0\nsTKX299rIwKBgH/gB3iSNFBhgzBe90LS7iYnxk8viMjQOi6MI4C2dbkXTCBuQxQ9\nywAjPp5htpXP3eAsQnXCwjsH6JNPlw/aMnDYqMM4/TD5OHiKCVWhPuGU9HY2A3KB\nkK/+HkL8GykJd4lDq5cOG9WUESg4Avqk4sNzSrKzODsL4RJmSt4MZHLJAoGBAI+K\ndgt6IFxVGkwYCDeQq1IHOvQnGlFTxgIw685pCQ/02IBRRHtJNyXsa/oObePWW7U5\nkivOEugE4MkO8vPv7T8V9KlTH/3IdYiPSqpLMJ5yjuBKIZ6/7Ua1Ol8FPBrdS7vJ\nrj+jGHdnrsSqPoCDPCTEq2ucHSDzLZM3Ci0+pUylAoGARDQaP2XOHQGiNg5DPM4f\nsWv1Yy4TOPQCUvkOwWstF5kfvEpLFj2TojcIPM5wAnTvwKLLHgOQhDon2R1sNhly\nNpLulowr9Z1bGPlNhG11698uVuabUMdX0Wl/m6YvqzbUKe6DHE8x8ew2UxMcEHyG\n/+jiYBZzOvAvHWhIhRtaywE=\n-----END PRIVATE KEY-----\n"
	Wechat_ApiCert  = "-----BEGIN CERTIFICATE-----\nMIID8DCCAtigAwIBAgIULgF6rdp/fzmZT4Q4m19A3KGVBw4wDQYJKoZIhvcNAQEL\nBQAwXjELMAkGA1UEBhMCQ04xEzARBgNVBAoTClRlbnBheS5jb20xHTAbBgNVBAsT\nFFRlbnBheS5jb20gQ0EgQ2VudGVyMRswGQYDVQQDExJUZW5wYXkuY29tIFJvb3Qg\nQ0EwHhcNMjEwNDIzMTA1MDQ0WhcNMjYwNDIyMTA1MDQ0WjCBgTETMBEGA1UEAwwK\nMTYwODQ2MDY2MjEbMBkGA1UECgwS5b6u5L+h5ZWG5oi357O757ufMS0wKwYDVQQL\nDCTlub/lt57ov57nsbPkv6Hmga/np5HmioDmnInpmZDlhazlj7gxCzAJBgNVBAYM\nAkNOMREwDwYDVQQHDAhTaGVuWmhlbjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC\nAQoCggEBALAezGDFFY5IC6BY5FSXsuFGUZoNzLRTKmWwNUBp9ZfuO5wcdidpDTPl\ntzuge3bEzFhaKudChVEp8rSgdfFyUF8z6bf7Qt3+QRSr3YAC+TZfIwXU48/xzxvY\nP31vv6pisOzBmRKoPWKnxTpBWPfLLgwjyZdQsJRyJ6XU7+j1kRGD09GecLwt5mYc\n9hCfTGWxLZTLJmuPCl93ANDNlan6kOCgzj8kPqeDz4w54IiCmSu/uPsyUlXlGGw8\ndrEYHWZ6oVIU4XNFiPua8qsIvJ0k8epbuybGVKa6kvZ4OSHqmnMa0r0MqWv+CWIf\nHyjEcXheYPneovrzP3pZRMaxazKnw9cCAwEAAaOBgTB/MAkGA1UdEwQCMAAwCwYD\nVR0PBAQDAgTwMGUGA1UdHwReMFwwWqBYoFaGVGh0dHA6Ly9ldmNhLml0cnVzLmNv\nbS5jbi9wdWJsaWMvaXRydXNjcmw/Q0E9MUJENDIyMEU1MERCQzA0QjA2QUQzOTc1\nNDk4NDZDMDFDM0U4RUJEMjANBgkqhkiG9w0BAQsFAAOCAQEAgQ8b1W9vW0LZMpDO\nIGjDNXplgVT0iCa9dXO0o4yeP5FLk3odp8mikZzsNsVIaovww918ybiknyZ1NiPz\nn/0XUNjWcQeTqX4+0R0+61isSQVtraXm2MrplfBG+5F+7mAsSgtIOcgMYb6SUHl1\ntLiJvl89FlMZuwJ7eIFA1FCw+7w82Uy0/PevthLfWqLNbMV9aiD5g4NHT2xTcJpx\natTcstRbjsHCV1WwOohijWf2bNGf+Cfntbw59LtXt1PY5mO0bI0bCz/Pja/75OTT\nc8+1PfvlXR4vEsJkaaDVWX3mexVlFO0EeGiuUstX6itkTn6G/im3t3wjI0tvPm6e\nBM04rA==\n-----END CERTIFICATE-----\n"
)

func main() {

	client := wechat.NewClient(WechatPay_appID, WechatPay_mchId, WechatPay_apiKey, true)

	eerr := client.AddCertPemFilePath("./configs/apiclient_cert.pem", "./configs/apiclient_key.pem")

	if eerr != nil {
		fmt.Println("fail ", eerr)
		return
	}
	// 打开Debug开关，输出日志
	client.DebugSwitch = gopay.DebugOn

	// 设置国家，不设置默认就是 China
	client.SetCountry(wechat.China)

	type Receiver struct {
		Type        string `json:"type"`
		Account     string `json:"account"`
		Amount      int    `json:"amount"`
		Description string `json:"description"`
	}

	// 初始化参数结构体
	bm := make(gopay.BodyMap)
	bm.Set("nonce_str", util.GetRandomString(32)).
		Set("transaction_id", "4200001000202104297029667273").
		// Set("sub_mch_id", WechatPay_SUBmchId).
		// Set("sub_appid", WechatPay_SUBAppid).
		// Set("appid", WechatPay_appID).
		// Set("mch_id", WechatPay_mchId).
		Set("out_order_no", "P20210426123459")

	var rs []*Receiver
	item := &Receiver{
		Type:        "MERCHANT_ID",
		Account:     WechatPay_SUBmchId_LM,
		Amount:      1,
		Description: "test",
	}
	rs = append(rs, item)
	//item2 := &Receiver{
	//	Type:        "MERCHANT_ID",
	//	Account:     WechatPay_SUBmchId,
	//	Amount:      2,
	//	//Description: "分到子商户",
	//	Description: "test2",
	//}
	//rs = append(rs, item2)
	bs, _ := json.Marshal(rs)

	_ = bs
	bm.Set("receivers", string(bs))
	//bm.Set("receivers", string(bs))

	wxRsp, err := client.ProfitSharing(bm)
	if err != nil {
		xlog.Errorf("client.ProfitSharingAddReceiver(%+v),error:%+v", bm, err)
		return
	}
	xlog.Debug("wxRsp:", wxRsp)

}
