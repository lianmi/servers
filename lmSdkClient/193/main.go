//package main
//
//import (
//	"crypto/tls"
//	"crypto/x509"
//	"fmt"
//	mqtt "github.com/eclipse/paho.mqtt.golang"
//	"io/ioutil"
//	"log"
//	"time"
//)
//
//func NewTlsConfig() *tls.Config {
//	certpool := x509.NewCertPool()
//	ca, err := ioutil.ReadFile("ca/ca.crt")
//	if err != nil {
//		log.Fatalln(err.Error())
//	}
//	certpool.AppendCertsFromPEM(ca)
//	// Import client certificate/key pair
//	clientKeyPair, err := tls.LoadX509KeyPair("ca/mqtt.lianmi.cloud.crt", "ca/mqtt.lianmi.cloud.key")
//	if err != nil {
//		panic(err)
//	}
//	return &tls.Config{
//		RootCAs: certpool,
//		ClientAuth: tls.NoClientCert,
//		ClientCAs: nil,
//		InsecureSkipVerify: true,
//		Certificates: []tls.Certificate{clientKeyPair},
//	}
//}
//
//func sub(client mqtt.Client) {
//	topic := "topic/test"
//	token := client.Subscribe(topic, 1, nil)
//	token.Wait()
//	fmt.Printf("Subscribed to topic %s", topic)
//}
//
//func publish(client mqtt.Client) {
//	num := 10
//	for i := 0; i < num; i++ {
//		text := fmt.Sprintf("Message %d", i)
//		token := client.Publish("topic/test", 0, false, text)
//		token.Wait()
//		time.Sleep(time.Second)
//	}
//}
//func main() {
//	opts := mqtt.NewClientOptions().AddBroker("ssl://192.168.1.193:1883").SetClientID("sample")
//	opts.SetProtocolVersion(4)
//	tlsConfig := NewTlsConfig()
//
//	opts.SetTLSConfig(tlsConfig)
//
//
//	c := mqtt.NewClient(opts)
//	if token := c.Connect(); token.Wait() && token.Error() != nil {
//		panic(token.Error())
//	}
//	sub(c)
//	publish(c)
//
//	c.Disconnect(100000)
//
//}

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"io/ioutil"
	"log"
	"time"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	var broker = "192.168.1.193"
	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("ssl://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client")
	//opts.SetUsername("emqx")
	//opts.SetPassword("public")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	tlsConfig:= NewTlsConfig()
	opts.SetTLSConfig(tlsConfig)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	sub(client)
	publish(client)

	client.Disconnect(250)
}

func publish(client mqtt.Client) {
	num := 10
	for i := 0; i < num; i++ {
		text := fmt.Sprintf("Message %d", i)
		token := client.Publish("test", 0, false, text)
		token.Wait()
		time.Sleep(time.Second)
	}
}

func sub(client mqtt.Client) {
	topic := "test"
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s", topic)
}

func NewTlsConfig() *tls.Config {
	certpool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("./ca.crt")
	if err != nil {
		log.Fatalln(err.Error())
	}
	certpool.AppendCertsFromPEM(ca)
	// Import client certificate/key pair
	//clientKeyPair, err := tls.LoadX509KeyPair("ca/mqtt.lianmi.cloud.crt", "ca/mqtt.lianmi.cloud.key")
	clientKeyPair, err := tls.LoadX509KeyPair("./192.168.1.193.crt", "./192.168.1.193.key")
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		RootCAs: certpool,
		ClientAuth: tls.NoClientCert,
		ClientCAs: nil,
		InsecureSkipVerify: true,
		Certificates: []tls.Certificate{clientKeyPair},
	}
}