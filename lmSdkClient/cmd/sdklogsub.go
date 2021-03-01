/*

 */
package cmd

import (
	"github.com/spf13/cobra"

	"context"

	"github.com/lianmi/servers/lmSdkClient/business"

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
				log.Println(payload)

			}
		}

	},
}

func init() {
	rootCmd.AddCommand(sdklogsubCmd)

	sdklogsubCmd.PersistentFlags().StringP("username", "u", "id1", "register username ")
}
