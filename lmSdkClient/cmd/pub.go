/*
mqtt 测试 之 pub
*/
package cmd

import (
	"github.com/spf13/cobra"

	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/eclipse/paho.golang/paho"
	"io"
	"io/ioutil"
	"log"
	// "net"
	"os"
	"os/signal"
	"syscall"
)

const (
	// localDeviceID = "lishijia-golang"
	CaPath = "/Users/mac/developments/lianmi/lm-cloud/servers/lmSdkClient/ca"
)

func NewTlsConfig() *tls.Config {
	certpool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(CaPath + "/ca.crt")
	if err != nil {
		log.Fatalln(err.Error())
	} else {
		log.Println("ReadFile ok")
	}
	certpool.AppendCertsFromPEM(ca)
	clientKeyPair, err := tls.LoadX509KeyPair(CaPath+"/mqtt.lianmi.cloud.crt", CaPath+"/mqtt.lianmi.cloud.key")
	if err != nil {
		panic(err)
	} else {
		log.Println("LoadX509KeyPair ok")
	}
	return &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{clientKeyPair},
	}
}

// pubCmd represents the pub command
var pubCmd = &cobra.Command{
	Use:   "pub",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("pub called")
		stdin := bufio.NewReader(os.Stdin)

		server, _ := cmd.PersistentFlags().GetString("server")
		topic, _ := cmd.PersistentFlags().GetString("topic")
		qos, _ := cmd.PersistentFlags().GetInt("qos")
		retained, _ := cmd.PersistentFlags().GetBool("retained")
		clientid, _ := cmd.PersistentFlags().GetString("clientid")

		// conn, err := net.Dial("tcp", *server)
		tlsConfig := NewTlsConfig()
		// broker := fmt.Sprintf("%s:%d", "192.168.1.193", 1883)
		conn, err := tls.Dial("tcp", server, tlsConfig)
		if err != nil {
			log.Fatalf("Failed to connect to %s: %s", server, err)
		}

		c := paho.NewClient(paho.ClientConfig{
			Conn: conn,
		})

		cp := &paho.Connect{
			KeepAlive:  30,
			ClientID:   clientid,
			CleanStart: true,
			// Username:   "",
			// Password:   ,
		}

		log.Println(cp.UsernameFlag, cp.PasswordFlag)

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

		for {
			message, err := stdin.ReadString('\n')
			if err == io.EOF {
				os.Exit(0)
			}

			if _, err = c.Publish(context.Background(), &paho.Publish{
				Topic:   topic,
				QoS:     byte(qos),
				Retain:  retained,
				Payload: []byte(message),
			}); err != nil {
				log.Println("error sending message:", err)
				continue
			}
			log.Println("sent")
		}
	},
}

func init() {
	hostname, _ := os.Hostname()
	//子命令
	mqttCmd.AddCommand(pubCmd)
	pubCmd.PersistentFlags().StringP("server", "s", "mqtt.lianmi.cloud:1883", "The full URL of the MQTT server to connect to")
	pubCmd.PersistentFlags().StringP("topic", "t", hostname, "Topic to publish the messages on")
	pubCmd.PersistentFlags().IntP("qos", "q", 1, "qos")
	pubCmd.PersistentFlags().BoolP("retained", "r", false, "retained")
	pubCmd.PersistentFlags().StringP("clientid", "c", "", "clientid")

}
