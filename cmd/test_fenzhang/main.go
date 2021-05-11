package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/iGoogle-ink/gopay"
	"github.com/iGoogle-ink/gopay/pkg/util"
	"github.com/iGoogle-ink/gopay/pkg/xlog"
	"github.com/iGoogle-ink/gopay/wechat"
	//"test_wx_pay/wechat"
)

const (
	WechatPay_appID    = "wx9dff85e1c3a3b342"
	WechatPay_apiV3Key = "LianmiLianmiLianmiLianmicloud508" //v3
	WechatPay_mchId    = "1608460662"
	//WechatPay_mchId            = "1608720021"
	//WechatPay_SUBmchId            = "1608720021"

	WechatPay_SUBmchId         = "1608737479"
	WechatPay_SUBmchId_LM      = "1608720021"
	WechatPay_SUBAppid         = "wx239c33b9be7cd047"
	WechatPay_SUBAppid_LM      = "wx9dff85e1c3a3b342"
	WechatPay_platformCertPath = "./configs/apiclient_cert.p12"
)

func PrintPretty(i interface{}) {
	data, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	fmt.Printf("%s\n", data)
}

func main() {

	client := wechat.NewClient(WechatPay_appID, WechatPay_mchId, WechatPay_apiV3Key, true)

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
		Set("sub_mch_id", WechatPay_SUBmchId).
		// Set("sub_appid", WechatPay_SUBAppid). 不需要
		// Set("appid", WechatPay_appID). 不需要
		// Set("mch_id", WechatPay_mchId). 不需要
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
	xlog.Debug("wxRsp:")
	PrintPretty(wxRsp)

}
