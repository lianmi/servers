/*
mqtt tls test
*/
package cmd

import (
	// "fmt"

	"github.com/spf13/cobra"
	// "context"
	// "errors"
	// "net"
	// "crypto/tls"
	// "crypto/x509"
	// "io/ioutil"
	// "time"
	// "github.com/golang/protobuf/proto"
	// User "github.com/lianmi/servers/api/proto/user"
	// "github.com/lianmi/servers/lmSdkClient/common"
	// "log"
	// "github.com/eclipse/paho.golang/paho" //支持v5.0
)

// const (
// 	localDeviceID = "lishijia-golang"
// )

// func NewTlsConfig() *tls.Config {
// 	certpool := x509.NewCertPool()
// 	ca, err := ioutil.ReadFile(common.CaPath + "/ca.crt")
// 	if err != nil {
// 		log.Fatalln(err.Error())
// 	} else {
// 		log.Println("ReadFile ok")
// 	}
// 	certpool.AppendCertsFromPEM(ca)
// 	clientKeyPair, err := tls.LoadX509KeyPair(common.CaPath+"/192.168.1.193.crt", common.CaPath+"/192.168.1.193.key")
// 	if err != nil {
// 		panic(err)
// 	} else {
// 		log.Println("LoadX509KeyPair ok")
// 	}
// 	return &tls.Config{
// 		RootCAs:            certpool,
// 		ClientAuth:         tls.NoClientCert,
// 		ClientCAs:          nil,
// 		InsecureSkipVerify: true,
// 		Certificates:       []tls.Certificate{clientKeyPair},
// 	}
// }

// mqttCmd represents the mqtt command
var mqttCmd = &cobra.Command{
	Use:   "mqtt",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		//
	},
}

func init() {
	rootCmd.AddCommand(mqttCmd)
}
