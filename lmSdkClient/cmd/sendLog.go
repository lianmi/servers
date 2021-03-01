/*
topic: lianmi/cloud/sdklogs
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/eclipse/paho.golang/paho"
	"github.com/lianmi/servers/lmSdkClient/business"
	"github.com/spf13/cobra"
)

// sendLogCmd represents the sendLog command
var sendLogCmd = &cobra.Command{
	Use:   "sendLog",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sendLog called")
		content, _ := cmd.PersistentFlags().GetString("content")

		topic := "lianmi/cloud/dispatcher"

		pb := &paho.Publish{
			Topic:   topic,
			QoS:     byte(2),
			Payload: []byte(content),
			Properties: &paho.PublishProperties{
				User: map[string]string{
					"jwtToken":        "",   // jwt令牌
					"deviceId":        "",   // 设备号
					"businessType":    "98", // 日志专用业务号
					"businessSubType": "0",  // 业务子号
					"taskId":          "0",
					"code":            "0",
					"errormsg":        "",
				},
			},
		}

		var client *paho.Client
		var payloadCh chan []byte
		payloadCh = make(chan []byte, 0)

		client = business.CreateClient(payloadCh)

		cp := &paho.Connect{
			KeepAlive:  30,
			ClientID:   "id1",
			CleanStart: true,
			Username:   "",
			Password:   []byte(""),
		}
		_, err := client.Connect(context.Background(), cp)
		if err == nil {
			log.Println("Succeed to Connect mqtt server")
		} else {
			log.Println("Failed to Connect mqtt server", err)
		}

		if _, err := client.Publish(context.Background(), pb); err != nil {
			log.Println("Failed to Publish:", err)
		} else {
			log.Println("Succeed Publish to mqtt broker:", topic)
		}
	},
}

func init() {
	rootCmd.AddCommand(sendLogCmd)
	sendLogCmd.PersistentFlags().StringP("content", "c", "", "content")
}
