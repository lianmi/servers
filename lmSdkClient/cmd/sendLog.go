/*
topic: lianmi/cloud/sdklogs
*/
package cmd

import (
	"context"
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
		content, _ := cmd.PersistentFlags().GetString("content")

		topic := "lianmi/cloud/dispatcher"

		props := &paho.PublishProperties{}
		// props.ResponseTopic = responseTopic
		props.User = props.User.Add("jwtToken", "")
		props.User = props.User.Add("deviceId", "")
		props.User = props.User.Add("businessType", "11")
		props.User = props.User.Add("businessSubType", "1")
		props.User = props.User.Add("taskId", "1")
		props.User = props.User.Add("code", "0")
		props.User = props.User.Add("errormsg", "")

		pb := &paho.Publish{
			Topic:      topic,
			QoS:        byte(2),
			Payload:    []byte(content),
			Properties: props,
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
