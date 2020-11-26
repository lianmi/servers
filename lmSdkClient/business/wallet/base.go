package wallet

import (
	// "context"
	"crypto/tls"
	"crypto/x509"
	// "errors"
	// "fmt"

	// "github.com/eclipse/paho.golang/paho" //支持v5.0

	clientcommon "github.com/lianmi/servers/lmSdkClient/common"
	"io/ioutil"
	"log"
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
