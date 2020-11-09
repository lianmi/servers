/*

 */
package cmd

import (
	"github.com/spf13/cobra"

	"context"
	"crypto/tls"
	"fmt"
	"github.com/eclipse/paho.golang/paho"
	"log"
	// "net"
	"os"
	"os/signal"
	"syscall"
)

// subCmd represents the sub command
var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("sub called")

		logger := log.New(os.Stdout, "SUB: ", log.LstdFlags)

		msgChan := make(chan *paho.Publish)

		server, _ := cmd.PersistentFlags().GetString("server")
		topic := "#"
		qos, _ := cmd.PersistentFlags().GetInt("qos")
		clientid, _ := cmd.PersistentFlags().GetString("clientid")

		tlsConfig := NewTlsConfig()
		conn, err := tls.Dial("tcp", server, tlsConfig)
		if err != nil {
			log.Fatalf("Failed to connect to %s: %s", server, err)
		}

		c := paho.NewClient(paho.ClientConfig{
			Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
				msgChan <- m
			}),
			Conn: conn,
		})
		c.SetDebugLogger(logger)
		c.SetErrorLogger(logger)

		cp := &paho.Connect{
			KeepAlive:  30,
			ClientID:   clientid,
			CleanStart: true,
		}

		ca, err := c.Connect(context.Background(), cp)
		if err != nil {
			log.Fatalln(err)
		}
		if ca.ReasonCode != 0 {
			log.Fatalf("Failed to connect to %s : %d - %s", server, ca.ReasonCode, ca.Properties.ReasonString)
		}

		fmt.Printf("Connected to %s\n", server)

		ic := make(chan os.Signal, 1)
		signal.Notify(ic, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-ic
			fmt.Println("signal received, exiting")
			if c != nil {
				d := &paho.Disconnect{ReasonCode: 0}
				c.Disconnect(d)
			}
			os.Exit(0)
		}()

		sa, err := c.Subscribe(context.Background(), &paho.Subscribe{
			Subscriptions: map[string]paho.SubscribeOptions{
				topic: {QoS: byte(qos)},
			},
		})
		if err != nil {
			log.Fatalln(err)
		}
		if sa.Reasons[0] != byte(qos) {
			log.Fatalf("Failed to subscribe to %s : %d", topic, sa.Reasons[0])
		}
		log.Printf("Subscribed to %s", topic)

		for m := range msgChan {
			log.Println("Received message:", string(m.Payload))
		}

	},
}

func init() {
	//子命令
	mqttCmd.AddCommand(subCmd)
	subCmd.PersistentFlags().StringP("server", "s", "mqtt.lianmi.cloud:1883", "The full URL of the MQTT server to connect to")
	subCmd.PersistentFlags().IntP("qos", "q", 1, "qos")
	subCmd.PersistentFlags().StringP("clientid", "c", "", "clientid")
}
