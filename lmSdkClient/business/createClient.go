package business

import (
	// "context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"

	// "errors"
	// "fmt"
	"io/ioutil"
	"log"

	// "time"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	// "github.com/golang/protobuf/proto"
	// "github.com/gomodule/redigo/redis"

	// Msg "github.com/lianmi/servers/api/proto/msg"
	"net"

	LMCommon "github.com/lianmi/servers/internal/common"
	clientcommon "github.com/lianmi/servers/lmSdkClient/common"
)

func NewTlsConfig() *tls.Config {
	certpool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(clientcommon.CaPath + "/ca.crt")
	if err != nil {
		log.Fatalln(err.Error())
	} else {
		log.Println("ReadFile ok")
	}
	certpool.AppendCertsFromPEM(ca)
	clientKeyPair, err := tls.LoadX509KeyPair(clientcommon.CaPath+"/mqtt.lianmi.cloud.crt", clientcommon.CaPath+"/mqtt.lianmi.cloud.key")
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
func PrintPretty(i interface{}) {
	data, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	fmt.Printf("%s\n", data)
}

func CreateClient(payloadCh chan []byte) *paho.Client {
	//使用ca
	var client *paho.Client
	if LMCommon.IsUseCa {
		//Connect mqtt broker using ssl
		tlsConfig := NewTlsConfig()
		conn, err := tls.Dial("tcp", clientcommon.BrokerAddr, tlsConfig)
		if err != nil {
			log.Fatalf("Failed to connect to %s: %s", clientcommon.BrokerAddr, err)
		}

		// Create paho client.
		client = paho.NewClient(paho.ClientConfig{
			Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
				log.Println("Incoming mqtt broker message")
				log.Println("m.Properties.User 长度", len(m.Properties.User))

				PrintPretty(m.Properties)

				topic := m.Topic
				jwtToken := m.Properties.User.Get("jwtToken") // Add by lishijia  for flutter mqtt
				deviceId := m.Properties.User.Get("deviceId")
				businessTypeStr := m.Properties.User.Get("businessType")
				businessSubTypeStr := m.Properties.User.Get("businessSubType")
				taskIdStr := m.Properties.User.Get("taskId")
				code := m.Properties.User.Get("code")

				log.Println("topic: ", topic)
				log.Println("jwtToken: ", jwtToken)
				log.Println("deviceId: ", deviceId)
				log.Println("businessType: ", businessTypeStr)
				log.Println("businessSubType: ", businessSubTypeStr)
				log.Println("taskId: ", taskIdStr)
				log.Println("code: ", code)

				if code == "200" {
					log.Println("Response succeed")
					// 回包
					payloadCh <- m.Payload

				} else {
					log.Println("Response failed")
				}

			}),
			Conn: conn,
		})

	} else {
		conn, err := net.Dial("tcp", clientcommon.BrokerAddr)
		if err != nil {
			log.Fatalf("Failed to connect to %s: %s", clientcommon.BrokerAddr, err)
		}

		// Create paho client.
		client = paho.NewClient(paho.ClientConfig{
			Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
				log.Println("Incoming mqtt broker message")

				topic := m.Topic
				jwtToken := m.Properties.User.Get("jwtToken") // Add by lishijia  for flutter mqtt
				deviceId := m.Properties.User.Get("deviceId")
				businessTypeStr := m.Properties.User.Get("businessType")
				businessSubTypeStr := m.Properties.User.Get("businessSubType")
				taskIdStr := m.Properties.User.Get("taskId")
				code := m.Properties.User.Get("code")
				errormsg := m.Properties.User.Get("errormsg")

				log.Println("topic: ", topic)
				log.Println("jwtToken: ", jwtToken)
				log.Println("deviceId: ", deviceId)
				log.Println("businessType: ", businessTypeStr)
				log.Println("businessSubType: ", businessSubTypeStr)
				log.Println("taskId: ", taskIdStr)

				if code == "200" {
					// 回包
					log.Println("Response succeed")
					payloadCh <- m.Payload

				} else {
					log.Printf("Request failed, code: %d, errormsg: %s\n", code, errormsg)
				}

			}),
			Conn: conn,
		})
	}

	return client
}
