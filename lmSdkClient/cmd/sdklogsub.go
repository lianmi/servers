/*

 */
package cmd

import (
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"

	"context"

	"github.com/lianmi/servers/lmSdkClient/business"
	Log "github.com/lianmi/servers/api/proto/log"

	"fmt"
	"log"

	"github.com/eclipse/paho.golang/paho"
)

// sdklogsubCmd represents the sdklogsub command
var sdklogsubCmd = &cobra.Command{
	Use:   "sdklogsub",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sdklogsub called")

		var client *paho.Client
		var payloadCh chan []byte
		payloadCh = make(chan []byte, 0)

		client = business.CreateClient(payloadCh)

		cp := &paho.Connect{
			KeepAlive:  30,
			ClientID:   "sdklogs",
			CleanStart: true,
			Username:   "",
			Password:   []byte(""),
		}
		ca, err := client.Connect(context.Background(), cp)
		if err == nil {
			if ca.ReasonCode == 0 {
				subTopic := fmt.Sprintf("lianmi/cloud/sdklogs")
				if _, err := client.Subscribe(context.Background(), &paho.Subscribe{
					Subscriptions: map[string]paho.SubscribeOptions{
						subTopic: paho.SubscribeOptions{QoS: byte(2), NoLocal: true},
					},
				}); err != nil {
					log.Println("Failed to subscribe:", err)
				}
				log.Println("Subscribed succed: ", subTopic)
			}
		} else {
			log.Println("Failed to Connect mqtt server", err)
		}

		//堵塞

		for {
			select {
			case payload := <-payloadCh:
				// log.Println(payload)

				//解包负载 payload
				var rsq Log.SendLogReq
				if err := proto.Unmarshal(payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {

					log.Println("回包内容---------------------")
					// log.Println("blockNumber: ", rsq.BlockNumber)
					// log.Println("hash: ", rsq.Hash)
					// log.Println("AmountLNMC: ", rsq.AmountLNMC)
					// log.Println("Time: ", rsq.Time)

				}

			}
		}

	},
}

func init() {
	rootCmd.AddCommand(sdklogsubCmd)

	sdklogsubCmd.PersistentFlags().StringP("username", "u", "id1", "register username ")
}
