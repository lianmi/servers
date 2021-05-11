package main

import (
	"encoding/json"
	"fmt"
	"log"

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
	// Wechat_ApiCert             = "-----BEGIN CERTIFICATE-----\nMIID8DCCAtigAwIBAgIULgF6rdp/fzmZT4Q4m19A3KGVBw4wDQYJKoZIhvcNAQEL\nBQAwXjELMAkGA1UEBhMCQ04xEzARBgNVBAoTClRlbnBheS5jb20xHTAbBgNVBAsT\nFFRlbnBheS5jb20gQ0EgQ2VudGVyMRswGQYDVQQDExJUZW5wYXkuY29tIFJvb3Qg\nQ0EwHhcNMjEwNDIzMTA1MDQ0WhcNMjYwNDIyMTA1MDQ0WjCBgTETMBEGA1UEAwwK\nMTYwODQ2MDY2MjEbMBkGA1UECgwS5b6u5L+h5ZWG5oi357O757ufMS0wKwYDVQQL\nDCTlub/lt57ov57nsbPkv6Hmga/np5HmioDmnInpmZDlhazlj7gxCzAJBgNVBAYM\nAkNOMREwDwYDVQQHDAhTaGVuWmhlbjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC\nAQoCggEBALAezGDFFY5IC6BY5FSXsuFGUZoNzLRTKmWwNUBp9ZfuO5wcdidpDTPl\ntzuge3bEzFhaKudChVEp8rSgdfFyUF8z6bf7Qt3+QRSr3YAC+TZfIwXU48/xzxvY\nP31vv6pisOzBmRKoPWKnxTpBWPfLLgwjyZdQsJRyJ6XU7+j1kRGD09GecLwt5mYc\n9hCfTGWxLZTLJmuPCl93ANDNlan6kOCgzj8kPqeDz4w54IiCmSu/uPsyUlXlGGw8\ndrEYHWZ6oVIU4XNFiPua8qsIvJ0k8epbuybGVKa6kvZ4OSHqmnMa0r0MqWv+CWIf\nHyjEcXheYPneovrzP3pZRMaxazKnw9cCAwEAAaOBgTB/MAkGA1UdEwQCMAAwCwYD\nVR0PBAQDAgTwMGUGA1UdHwReMFwwWqBYoFaGVGh0dHA6Ly9ldmNhLml0cnVzLmNv\nbS5jbi9wdWJsaWMvaXRydXNjcmw/Q0E9MUJENDIyMEU1MERCQzA0QjA2QUQzOTc1\nNDk4NDZDMDFDM0U4RUJEMjANBgkqhkiG9w0BAQsFAAOCAQEAgQ8b1W9vW0LZMpDO\nIGjDNXplgVT0iCa9dXO0o4yeP5FLk3odp8mikZzsNsVIaovww918ybiknyZ1NiPz\nn/0XUNjWcQeTqX4+0R0+61isSQVtraXm2MrplfBG+5F+7mAsSgtIOcgMYb6SUHl1\ntLiJvl89FlMZuwJ7eIFA1FCw+7w82Uy0/PevthLfWqLNbMV9aiD5g4NHT2xTcJpx\natTcstRbjsHCV1WwOohijWf2bNGf+Cfntbw59LtXt1PY5mO0bI0bCz/Pja/75OTT\nc8+1PfvlXR4vEsJkaaDVWX3mexVlFO0EeGiuUstX6itkTn6G/im3t3wjI0tvPm6e\nBM04rA==\n-----END CERTIFICATE-----\n"

	WechatPay_serierNo = "7D44E512E73027719552E38F0DE879D1A76C2B87" // 服务商证书序列号，更换证书后需要修改
	WechatPay_apiV3Key = "LianmiLianmiLianmiLianmicloud508"         // 服务商 apikey
	Wechat_pkContent   = `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDMoaqW4aXuuRcR
24TSzoUCbzGXf2DP/aI8dFoXH3/kNF7H/GejfZ/TuvM3R5oOWsN0BXdmY1hdO4q3
sdiLtpSus6SCGi45iw/v+3JJa/u2pDqQbN4ZTNZlOSZjfrlAmUqG1G47cg7J/4p6
/RZEcEb7WtC2ETv/EE0Ge7lqUzfycJ/5EQTpTUg8mqcYcF3XC3GpR+uEaATc15zs
elJwPito7Th+fdrC2CaNQYvsxqTzjD6zaKnfJGTp6OccloGqn15bzoCpSsMSYFL/
NsjNM6Gfh9ANKmOe4MqY2o+6hreSUnSrjPllF9bIzR/yr9LraZdh/EumBjpF/Cb1
3MfdH3MpAgMBAAECggEBAIU/ZGi5aKZpSfdr3TK0HfJ223EOFcl6HBGHpj5WWZ4M
6AcLeaUBIXjqzIMbkdp1Cb7b7GL0n86d/fcdzKc1bd3QxnedeqonvmoDbukWcqL8
j9IJwhnxac4iB7hUBWdmKhxf6aO14qFwUAlEEiLghagY+70CvfGZ+L4XBKaSp+Sq
fG56dYpPC/Gch5BYf3pCStW9G/V9e6wFR5DGRNC52Svw2pMJ4pQcdHqfzKmlJYXR
cL8V8cXnxiTIiFDuYWiNAdEOeausc2UXJEg3cIfy9YcrQi2mT+twsVswoIYBmegT
P70XdjlVSPwZBGPIcEoTxPBPU+inRFz273l+pa5OKg0CgYEA+1Wl3fqRCpfcbQTR
ngPD8gXKJ2TXIuqlwEz1RKWNu/3tNpqQ7b/Hsn6Q0dEbRuxDX2R9O8jcZUHLVnQP
O/Q4tmHcuxWC1PdMavuI1U/TwrBebxT4W/Fduhc3aF30Vk8N5Y0nXRpccdus4vce
0qa0GdFmbPhaibXcxB/HAMKPEdMCgYEA0G4VaSSc/1RqeeYf57gO2FfXI3J0gdX4
d642PNvMxK/DsrTZlPeZrdNcOcvl670ubRNPpz6pfcdac9x0wgVIeU4ZqN2xGqlt
Pp+RFI2Ob9dV5/ANVO0aFIhFqeqTR8Bus7KSoTx2t75pvSq2VNGb7GL8mr3lfQH1
6FTkwRfdjZMCgYBl1LDMfG3xpc/IV/B6Hjpwv8nFJkVIP1wCyuuA8ba4WUyYGA3q
Vg6aEk+owxlTJfyyFKvs4hfx6rNxBrr5ZpznwETHhBKrKLtMiTdKffplYkIQraVm
0ydPc4KehZqusX8G56bwQPL9qqyklM1nOeW0pDPkqMc+DnIxAFMHysxewwKBgDOd
HxYzZ+FeoSNglkQGcz6lufPgMvO37diNPochkvqd3+NQH5VhHyBJd8wkLuKKrYV7
Q71Rqh0okcChNhSZxFGtwnLruyC0FgZs8ztYto4BkBdofZSrRksRV9b07NXW1FMR
hHgDBg8ISxz6B77HTUpjVNRo8/xZ0PBgnWknpMibAoGBAJUmMxGRFdO3WOGo80Od
kKVN98lq4ZUdA2zPYUMxkRDHS2u3aLshEA2vnseqKHabV8M9UXyvUK9uH8KT+8rn
3jE1PsaOyc+RPMC+jobPG8FJOZRYV5lDDAlLt/g8QKWUBx+jNaFQijVidWcjPjND
2G4qLkNFV/7SB+31YvVqwB7w
-----END PRIVATE KEY-----
`
)

func PrintPretty(i interface{}) {
	data, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	fmt.Printf("%s\n", data)
}

func main() {
	wxPayClient, err := wechat.NewClientV3(
		WechatPay_appID,
		WechatPay_mchId,
		WechatPay_serierNo,
		WechatPay_apiV3Key,
		Wechat_pkContent)

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

	fmt.Println("data:\n ", string(jsonData))
}
