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
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/eclipse/paho.golang/paho"
	"github.com/golang/protobuf/proto"
	User "github.com/lianmi/servers/api/proto/user"
)

const (
	// localDeviceID = "lishijia-golang"
	CaPath = "/Users/mac/developments/lianmi/lm-cloud/servers/lmSdkClient/ca"
)

//用于tls
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

		conn, err := net.Dial("tcp", server)
		/*
			tlsConfig := NewTlsConfig()
			conn, err := tls.Dial("tcp", server, tlsConfig)
		*/
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
			username, err := stdin.ReadString('\n')
			if err == io.EOF {
				os.Exit(0)
			}
			req := &User.GetUsersReq{}
			req.Usernames = append(req.Usernames, username)
			data, _ := proto.Marshal(req)

			jwtToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkZXZpY2VJRCI6ImJkMmQ1ZDBjLThiZjctNDY3Yy1iNTVjLWVhNWEwZTBmOGYwMyIsImV4cCI6MTYxMzM2MjIxNSwib3JpZ19pYXQiOjE2MTA3NzAyMTUsInVzZXJOYW1lIjoiaWQxIiwidXNlclJvbGVzIjoiW3tcImlkXCI6MSxcInVzZXJfaWRcIjoxLFwidXNlcl9uYW1lXCI6XCJpZDFcIixcInZhbHVlXCI6XCJcIn1dIn0.8ugMtx3l7S_6d21Y8yRCC-fAG1-IjWFOECkxrLYCKlk"

			pb := &paho.Publish{
				Topic:      topic,
				QoS:        byte(qos),
				Retain:     retained,
				Payload:    data,
				Properties: &paho.PublishProperties{
					// User: map[string]string{
					// 	"jwtToken":        jwtToken, // jwt令牌
					// 	"deviceId":        "b5d10669-403a-4e36-8b58-dbc31856126c",
					// 	"businessType":    "1",
					// 	"businessSubType": "1",
					// 	"taskId":          "1",
					// 	"code":            "200",
					// 	"errormsg":        "",
					// },
				},
			}

			pb.Properties.User.Add("jwtToken", jwtToken)
			pb.Properties.User.Add("deviceId", "b5d10669-403a-4e36-8b58-dbc31856126c")
			pb.Properties.User.Add("businessType", "1")
			pb.Properties.User.Add("businessSubType", "1")
			pb.Properties.User.Add("taskId", "1")
			pb.Properties.User.Add("code", "0")
			pb.Properties.User.Add("errormsg", "")

			if _, err = c.Publish(context.Background(), pb); err != nil {
				log.Println("error sending message:", err)
				continue
			}
			log.Println("sent")
		}
	},
}

func init() {
	// hostname, _ := os.Hostname()
	//子命令
	mqttCmd.AddCommand(pubCmd)
	pubCmd.PersistentFlags().StringP("server", "s", "127.0.0.1:1883", "The full URL of the MQTT server to connect to")
	pubCmd.PersistentFlags().StringP("topic", "t", "lianmi/cloud/dispatcher", "Topic to publish the messages on")
	pubCmd.PersistentFlags().IntP("qos", "q", 2, "qos")
	pubCmd.PersistentFlags().BoolP("retained", "r", false, "retained")
	pubCmd.PersistentFlags().StringP("clientid", "c", "", "clientid")

}
