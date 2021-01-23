/*
mqtt tls test
*/
package cmd

import (
	// "fmt"

	"github.com/spf13/cobra"
	// "context"
	// "errors"
	// "net"
	// "crypto/tls"
	// "crypto/x509"
	// "io/ioutil"
	// "time"
	// "github.com/golang/protobuf/proto"
	// User "github.com/lianmi/servers/api/proto/user"
	// "github.com/lianmi/servers/lmSdkClient/common"
	// "log"
	// "github.com/eclipse/paho.golang/paho" //支持v5.0
)

// const (
// 	localDeviceID = "lishijia-golang"
// )

// func NewTlsConfig() *tls.Config {
// 	certpool := x509.NewCertPool()
// 	ca, err := ioutil.ReadFile(common.CaPath + "/ca.crt")
// 	if err != nil {
// 		log.Fatalln(err.Error())
// 	} else {
// 		log.Println("ReadFile ok")
// 	}
// 	certpool.AppendCertsFromPEM(ca)
// 	clientKeyPair, err := tls.LoadX509KeyPair(common.CaPath+"/192.168.1.193.crt", common.CaPath+"/192.168.1.193.key")
// 	if err != nil {
// 		panic(err)
// 	} else {
// 		log.Println("LoadX509KeyPair ok")
// 	}
// 	return &tls.Config{
// 		RootCAs:            certpool,
// 		ClientAuth:         tls.NoClientCert,
// 		ClientCAs:          nil,
// 		InsecureSkipVerify: true,
// 		Certificates:       []tls.Certificate{clientKeyPair},
// 	}
// }

/*
//测试与flutter的mqtt _client 5.0通信
func mqttAct() {
	topic := "lianmi/cloud/dispatcher"
	localDeviceID := "localDeviceID-lishijia-0001"
	taskIdStr := fmt.Sprintf("%d", 1)

	pb := &paho.Publish{
		Topic: topic,
		QoS:   byte(1),
		// Payload: content,
		Properties: &paho.PublishProperties{
			// ResponseTopic:   responseTopic,
			User: map[string]string{
				"jwtToken":        jwtToken, // jwt令牌
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "10",          // 业务号
				"businessSubType": "1",           // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	if _, err := client.Publish(context.Background(), pb); err != nil {
		log.Println("Failed to Publish:", err)
	} else {
		log.Println("Succeed Publish to mqtt broker:", topic)
	}

	run := true
	ticker := time.NewTicker(30 * time.Second) // 30s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
}
*/

// mqttCmd represents the mqtt command
var mqttCmd = &cobra.Command{
	Use:   "mqtt",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		//
	},
}

func init() {
	rootCmd.AddCommand(mqttCmd)
}
