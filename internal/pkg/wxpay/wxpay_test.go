package wxpay

import (
	"fmt"
	"os"
	"testing"
)

var client *Client

func TestMain(m *testing.M) {
	client = New("wxf470593a8d5e0e12", "acf4990004d84488bd6cff67c0e15ade", "1604757586", true)

	// 加载退款需要的证书
	fmt.Println(client.LoadCert("/Users/mac/developments/xiaoma/cert/1604757586_20210113_cert/apiclient_cert.p12"))
	os.Exit(m.Run())
}
