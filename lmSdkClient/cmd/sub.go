/*

 */
package cmd

import (
	"github.com/spf13/cobra"

	"context"
	// "crypto/tls"

	"fmt"
	"github.com/eclipse/paho.golang/paho"
	"github.com/golang/protobuf/proto"
	User "github.com/lianmi/servers/api/proto/user"
	"log"
	"net"
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
		fmt.Println("sub called")

		logger := log.New(os.Stdout, "SUB: ", log.LstdFlags)

		msgChan := make(chan *paho.Publish)

		server, _ := cmd.PersistentFlags().GetString("server")
		topic := "lianmi/cloud/dispatcher"
		qos, _ := cmd.PersistentFlags().GetInt("qos")
		clientid, _ := cmd.PersistentFlags().GetString("clientid")

		/*
			tlsConfig := NewTlsConfig()
			conn, err := tls.Dial("tcp", server, tlsConfig)
			if err != nil {
				log.Fatalf("Failed to connect to %s: %s", server, err)
			}
		*/

		conn, err := net.Dial("tcp", server)
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
			// log.Println("Received message:", string(m.Payload))
			log.Println("=========================")
			topic := m.Topic
			log.Println("topic:", topic)

			jwtToken := m.Properties.User["jwtToken"]
			log.Println("jwtToken:", jwtToken)

			deviceId := m.Properties.User["deviceId"]
			log.Println("deviceId:", deviceId)

			businessTypeStr := m.Properties.User["businessType"]
			log.Println("businessType:", businessTypeStr)

			businessSubTypeStr := m.Properties.User["businessSubType"]
			log.Println("businessSubType:", businessSubTypeStr)

			taskIdStr := m.Properties.User["taskId"]
			log.Println("taskId:", taskIdStr)

			log.Println("Received message:", m.Payload)
			log.Println("=========================")
			log.Println()

			// jwtToken= "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkZXZpY2VJRCI6ImJkMmQ1ZDBjLThiZjctNDY3Yy1iNTVjLWVhNWEwZTBmOGYwMyIsImV4cCI6MTYxMzM2MjIxNSwib3JpZ19pYXQiOjE2MTA3NzAyMTUsInVzZXJOYW1lIjoiaWQxIiwidXNlclJvbGVzIjoiW3tcImlkXCI6MSxcInVzZXJfaWRcIjoxLFwidXNlcl9uYW1lXCI6XCJpZDFcIixcInZhbHVlXCI6XCJcIn1dIn0.8ugMtx3l7S_6d21Y8yRCC-fAG1-IjWFOECkxrLYCKlk"

			if businessTypeStr == "1" && businessSubTypeStr == "1" {
				//解包body
				// bodyData := []byte{10, 3, 105, 100, 52}
				// log.Println("bodyData:", bodyData)
				var getUsersReq User.GetUsersReq
				if err := proto.Unmarshal(m.Payload, &getUsersReq); err != nil {
					log.Println("Protobuf Unmarshal bodyData Error:", err)
					continue

				} else {
					for _, username := range getUsersReq.GetUsernames() {
						log.Println("username", username)
					}

					rsp := &User.GetUsersResp{
						Users: make([]*User.User, 0),
					}

					user := &User.User{
						Username: "id4",
						Nick:     "小吴哥哥",
						Mobile:   "15875317540",
						Avatar:   "https://zbj-bucket1.oss-cn-shenzhen.aliyuncs.com/avatar.JPG",
						// Gender:        userBaseData.GetGender(),
						// Label:         userBaseData.Label,
						UserType: User.UserType(1),
						State:    User.UserState(1),
						// ContactPerson: userBaseData.ContactPerson,
						Province: "广东省",
						City:     "广州市",
						County:   "天河区",
						Street:   "体育西路",
						Address:  "建和中心21楼 ",
					}

					rsp.Users = append(rsp.Users, user)
					// message := "1-1,  响应参数: GetUsersResp"
					data, _ := proto.Marshal(rsp)
					// _ = data

					if _, err = c.Publish(context.Background(), &paho.Publish{
						Topic:   "lianmi/cloud/device/testdeviceid",
						QoS:     byte(qos),
						Retain:  false,
						Payload: data,
						Properties: &paho.PublishProperties{
							User: map[string]string{
								"jwtToken":        jwtToken, // jwt令牌
								"deviceId":        "b5d10669-403a-4e36-8b58-dbc31856126c",
								"businessType":    "1",
								"businessSubType": "1",
								"taskId":          taskIdStr,
								"code":            "200",
								"errormsg":        "",
							},
						},
					}); err != nil {
						log.Println("error sending message:", err)
						continue
					}
					log.Println("1-1 sent to flutter ok")

				}

			}
		}

	},
}

func init() {
	//子命令
	mqttCmd.AddCommand(subCmd)
	subCmd.PersistentFlags().StringP("server", "s", "127.0.0.1:1883", "The full URL of the MQTT server to connect to")
	subCmd.PersistentFlags().IntP("qos", "q", 2, "qos") //默认是 2
	subCmd.PersistentFlags().StringP("clientid", "c", "", "clientid")
}
