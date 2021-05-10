package main

import (
	"encoding/json"
	"fmt"

	"github.com/iGoogle-ink/gopay"
	"github.com/iGoogle-ink/gopay/pkg/util"
	"github.com/iGoogle-ink/gopay/pkg/xlog"
	"github.com/iGoogle-ink/gopay/wechat/v3"
)

const (
	//WechatPay_appID  = "wx239c33b9be7cd047"
	WechatPay_appID  = "wx9dff85e1c3a3b342"
	WechatPay_apiKey = "LianmiLianmiLianmiLianmicloud508"
	WechatPay_mchId  = "1608460662"
	//WechatPay_mchId            = "1608720021"
	//WechatPay_SUBmchId            = "1608720021"

	WechatPay_SUBmchId         = "1608737479"
	WechatPay_SUBmchId_LM      = "1608720021"
	WechatPay_SUBmchId_LM2     = "1608460662"
	WechatPay_SUBAppid         = "wx9dff85e1c3a3b342"
	WechatPay_SUBAppid_LM      = "wx239c33b9be7cd047"
	WechatPay_platformCertPath = "./configs/apiclient_cert.p12"
	Wechat_apiV3Key            = "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCwHsxgxRWOSAug\nWORUl7LhRlGaDcy0UyplsDVAafWX7jucHHYnaQ0z5bc7oHt2xMxYWirnQoVRKfK0\noHXxclBfM+m3+0Ld/kEUq92AAvk2XyMF1OPP8c8b2D99b7+qYrDswZkSqD1ip8U6\nQVj3yy4MI8mXULCUciel1O/o9ZERg9PRnnC8LeZmHPYQn0xlsS2UyyZrjwpfdwDQ\nzZWp+pDgoM4/JD6ng8+MOeCIgpkrv7j7MlJV5RhsPHaxGB1meqFSFOFzRYj7mvKr\nCLydJPHqW7smxlSmupL2eDkh6ppzGtK9DKlr/gliHx8oxHF4XmD53qL68z96WUTG\nsWsyp8PXAgMBAAECggEAJEgj+GeBek8zPfQyDY82xZvT4bWoDxn26P306nEprAPi\n+dUPLi1BEAjpx3nXFW+TXTwuDHgyuLa4jidkRLo0/nfWVRWI/+yKAbUqK13xcxcE\nQwZJbCQ3c1gINFIaHQK5yfxUCXjpNHK1ebvOlTWhJdUViXuQ9PPTYSFNzyMPoJi9\nQRXr57VEpiPifFeKweFv6oHAX0G05Ng0IOrvhK0zq1Ca9xUohjuOO5EhMXugWFBk\nrmOK834ECEoDeAPujQIDvt6ShoUPPKb55whkcw+wEyhIDId/0Z5ByLb3NdwdW+xX\nMwX5CMB0rytdT9anbheg1xuYc0OHOCm3EYyyIb8dGQKBgQDYLtsQDuCfepNUDaCu\nMb7k+iAxYbH/xNm0GVGq0y1CS4B3kdH5sU4U/P5i34d9+cOW4f8xyXQDSiF9KwhH\nAvXdEtTMMuC5eZWIfTgQPFyGDfXnbJAfhQiHcp2thZuGeBzB6BLg1BPw8V76BlWT\nTUJ5rjvRBUNBDCNv+1U63FfZvQKBgQDQjvbIWMlzEuRGtR5z2o7VhwyERiQH2BYC\nbEzdjetHVFRYRpXWPODKhdbsXXEDMOi3dnel1CesIyW7AcJs9rAA7R8tXm/QGeYx\nimYZqfP0WWK2pNnDarvluacCXsDmbx5klAd0Ka59s8dcO6I17WNVPVSrQwg2fJp0\nsTKX299rIwKBgH/gB3iSNFBhgzBe90LS7iYnxk8viMjQOi6MI4C2dbkXTCBuQxQ9\nywAjPp5htpXP3eAsQnXCwjsH6JNPlw/aMnDYqMM4/TD5OHiKCVWhPuGU9HY2A3KB\nkK/+HkL8GykJd4lDq5cOG9WUESg4Avqk4sNzSrKzODsL4RJmSt4MZHLJAoGBAI+K\ndgt6IFxVGkwYCDeQq1IHOvQnGlFTxgIw685pCQ/02IBRRHtJNyXsa/oObePWW7U5\nkivOEugE4MkO8vPv7T8V9KlTH/3IdYiPSqpLMJ5yjuBKIZ6/7Ua1Ol8FPBrdS7vJ\nrj+jGHdnrsSqPoCDPCTEq2ucHSDzLZM3Ci0+pUylAoGARDQaP2XOHQGiNg5DPM4f\nsWv1Yy4TOPQCUvkOwWstF5kfvEpLFj2TojcIPM5wAnTvwKLLHgOQhDon2R1sNhly\nNpLulowr9Z1bGPlNhG11698uVuabUMdX0Wl/m6YvqzbUKe6DHE8x8ew2UxMcEHyG\n/+jiYBZzOvAvHWhIhRtaywE=\n-----END PRIVATE KEY-----\n"
	Wechat_ApiCert             = "-----BEGIN CERTIFICATE-----\nMIID8DCCAtigAwIBAgIULgF6rdp/fzmZT4Q4m19A3KGVBw4wDQYJKoZIhvcNAQEL\nBQAwXjELMAkGA1UEBhMCQ04xEzARBgNVBAoTClRlbnBheS5jb20xHTAbBgNVBAsT\nFFRlbnBheS5jb20gQ0EgQ2VudGVyMRswGQYDVQQDExJUZW5wYXkuY29tIFJvb3Qg\nQ0EwHhcNMjEwNDIzMTA1MDQ0WhcNMjYwNDIyMTA1MDQ0WjCBgTETMBEGA1UEAwwK\nMTYwODQ2MDY2MjEbMBkGA1UECgwS5b6u5L+h5ZWG5oi357O757ufMS0wKwYDVQQL\nDCTlub/lt57ov57nsbPkv6Hmga/np5HmioDmnInpmZDlhazlj7gxCzAJBgNVBAYM\nAkNOMREwDwYDVQQHDAhTaGVuWmhlbjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC\nAQoCggEBALAezGDFFY5IC6BY5FSXsuFGUZoNzLRTKmWwNUBp9ZfuO5wcdidpDTPl\ntzuge3bEzFhaKudChVEp8rSgdfFyUF8z6bf7Qt3+QRSr3YAC+TZfIwXU48/xzxvY\nP31vv6pisOzBmRKoPWKnxTpBWPfLLgwjyZdQsJRyJ6XU7+j1kRGD09GecLwt5mYc\n9hCfTGWxLZTLJmuPCl93ANDNlan6kOCgzj8kPqeDz4w54IiCmSu/uPsyUlXlGGw8\ndrEYHWZ6oVIU4XNFiPua8qsIvJ0k8epbuybGVKa6kvZ4OSHqmnMa0r0MqWv+CWIf\nHyjEcXheYPneovrzP3pZRMaxazKnw9cCAwEAAaOBgTB/MAkGA1UdEwQCMAAwCwYD\nVR0PBAQDAgTwMGUGA1UdHwReMFwwWqBYoFaGVGh0dHA6Ly9ldmNhLml0cnVzLmNv\nbS5jbi9wdWJsaWMvaXRydXNjcmw/Q0E9MUJENDIyMEU1MERCQzA0QjA2QUQzOTc1\nNDk4NDZDMDFDM0U4RUJEMjANBgkqhkiG9w0BAQsFAAOCAQEAgQ8b1W9vW0LZMpDO\nIGjDNXplgVT0iCa9dXO0o4yeP5FLk3odp8mikZzsNsVIaovww918ybiknyZ1NiPz\nn/0XUNjWcQeTqX4+0R0+61isSQVtraXm2MrplfBG+5F+7mAsSgtIOcgMYb6SUHl1\ntLiJvl89FlMZuwJ7eIFA1FCw+7w82Uy0/PevthLfWqLNbMV9aiD5g4NHT2xTcJpx\natTcstRbjsHCV1WwOohijWf2bNGf+Cfntbw59LtXt1PY5mO0bI0bCz/Pja/75OTT\nc8+1PfvlXR4vEsJkaaDVWX3mexVlFO0EeGiuUstX6itkTn6G/im3t3wjI0tvPm6e\nBM04rA==\n-----END CERTIFICATE-----\n"
)

func main() {
	wxPayClient, err := wechat.NewClientV3(WechatPay_appID, WechatPay_mchId, "2E017AADDA7F7F39994F84389B5F40DCA195070E", "LianmiLianmiLianmiLianmicloud508", Wechat_apiV3Key)

	if err != nil {
		fmt.Println("微信支付客户端初始化失败")
		return
	}

	wxPayClient.DebugSwitch = 1
	//if wxPayClient == nil {
	//	fmt.Println("微信支付客户端初始化失败")
	//	return
	//} else {
	//	//fmt.Println("微信支付客户端初始化成功")
	//	////err := wxPayClient.AddCertPkcs12FilePath(WechatPay_platformCertPath) //装载本地证书
	//	//err := wxPayClient.AddCertPemFilePath("configs/apiclient_cert.pem" , "configs/apiclient_key.pem") //装载本地证书
	//	//if err != nil {
	//	//	fmt.Println("微信支付证书加载失败")
	//	//	//fmt.Println(err)
	//	//	return
	//	//} else {
	//	//	//fmt.Println("")
	//	//	fmt.Println("LoadCert succeed")
	//	//}
	//}
	//var p = wxpay.UnifiedOrderParam{}
	//设置国家

	//wxPayClient.SetCountry(wechat.China)

	number := util.GetRandomString(32)
	xlog.Debug("out_trade_no:", number)

	//初始化参数Map
	bm := make(gopay.BodyMap)
	//bm.Set("nonce_str", util.GetRandomString(32)).
	//Set("body", "App支付测试").
	bm.Set("sp_appid", WechatPay_SUBAppid).
		Set("sp_mchid", WechatPay_mchId).
		Set("sub_appid", WechatPay_SUBAppid_LM).
		Set("sub_mchid", WechatPay_SUBmchId).
		Set("out_trade_no", number).
		Set("description", "测试商品2").
		//Set("total_fee", 1).
		//Set("spbill_create_ip", "127.0.0.1").
		//Set("notify_url", "https://api.lianmi.cloud/v1/callback/wechat").
		Set("notify_url", "http://paycallback.geejoan.cn/callback").
		//Set("trade_type", wechat.TradeType_H5).
		//Set("trade_type", wechat).
		//Set("device_info", "APP").
		SetBodyMap("amount", func(bmloc gopay.BodyMap) {
			bmloc.Set("total", 1).Set("currency", "CNY")
		}).
		//Set("sign_type", wechat.SignTypeRSA).
		SetBodyMap("settle_info", func(bmloc gopay.BodyMap) {
			bmloc.Set("profit_sharing", true)
		})
	//.
	//SetBodyMap("scene_info", func(bm gopay.BodyMap) {
	//	bm.SetBodyMap("h5_info", func(bm gopay.BodyMap) {
	//		bm.Set("type", "Wap")
	//		bm.Set("wap_url", "https://www.fumm.cc")
	//		bm.Set("wap_name", "H5测试支付")
	//	})
	//}) /*.Set("openid", "o0Df70H2Q0fY8JXh1aFPIRyOBgu8")*/

	///
	//请求支付下单，成功后得到结果
	wxRsp, err := wxPayClient.V3PartnerTransactionApp(bm)
	if err != nil {
		xlog.Error(err)
		return
	}
	xlog.Debug("Response：", wxRsp)

	//var p = wechat.UnifiedOrderResponse{}
	//p.Body = "微信充值"
	//p.ProductId = uuid.NewV4().String()
	//p.NotifyURL = "http://api.lianmi.cloud/v1/callback/wechat"
	//p.TradeType = wxpay.TradeTypeApp
	////p.SpbillCreateIP = context.ClientIP()
	////p.TotalFee = int(datainfo.TotalAmount * 100) // 单位1分钱
	//p.TotalFee = 1  // 单位1分钱
	//
	//p.SpAppid = "wx239c33b9be7cd047"
	//p.SpMchid = WechatPay_mchId
	//p.SubAppid = "wx239c33b9be7cd047"
	//p.SubMchid ="1608720021"
	//
	//OutTradeNo := fmt.Sprintf("%s%s%s%s%s", p.ProductId[0:8], p.ProductId[9:13], p.ProductId[14:18], p.ProductId[19:23], p.ProductId[24:36])
	//
	//p.OutTradeNo = OutTradeNo //orderinfo.OrderId // 后面增加渠道编号
	//
	////pc.logger.Debug("WalletRechargeByWeChat", zap.Any("Body", p))
	//fmt.Println("WalletRechargeByWeChat ")
	//result, err2 := wxPayClient.UnifiedOrder(p)
	//if err2 != nil {
	//	//fmt.Println("微信服务器返回错误：" + err2.Error())
	//	fmt.Println("微信服务器返回错误", zap.Error(err2))
	//	//common.RespFail(context, http.StatusNotFound, codes.ERROR, fmt.Sprintf("微信服务器返回错误 : %s", err2.Error()))
	//
	//	return
	//}
	//fmt.Println("微信服务器返回", zap.Any("result", result))
	//
	//var m = make(url.Values)
	//m.Set("appid", WechatPay_appID)
	//m.Set("partnerid", WechatPay_mchId)
	//m.Set("prepayid", result.PrepayId)
	//m.Set("noncestr", result.NonceStr)
	//m.Set("timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	//m.Set("package", "Sign=WXPay")
	//var sign = wxpay.SignMD5(m, WechatPay_apiKey)
	//m.Set("sign", sign)
	//var m = make(url.Values)
	//m.Set("appid", WechatPay_appID)
	//m.Set("partnerid", WechatPay_mchId)
	//m.Set("prepayid", wxRsp.Response.PrepayId)
	//m.Set("noncestr", wxRsp.SignInfo.HeaderNonce)
	//m.Set("timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	//m.Set("package", "Sign=WXPay")
	//var sign =wechat.V3VerifySign(m, WechatPay_apiKey)
	//m.Set("sign", sign)
	//
	//var re map[string]string = make(map[string]string, 0)
	//for k, v := range m {
	//	re[k] = v[0]
	//}
	//RespData(context, http.StatusOK, 200, re)

	//wxPayClient.Appid = WechatPay_SUBAppid_LM
	//wxPayClient.Mchid = WechatPay_SUBmchId_LM2
	app, err := wxPayClient.PaySignOfApp(wxRsp.Response.PrepayId)
	if err != nil {

		return
	}

	jsonData, err := json.Marshal(app)

	fmt.Println("data ", string(jsonData))
}
