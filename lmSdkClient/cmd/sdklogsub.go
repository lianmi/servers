/*

 */
package cmd

import (
	"io"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"

	"context"

	Log "github.com/lianmi/servers/api/proto/log"
	"github.com/lianmi/servers/lmSdkClient/business"

	"fmt"
	"log"

	"github.com/eclipse/paho.golang/paho"
)

const (
	//LOGPATH  LOGPATH/time.Now().Format(FORMAT)/*.log
	LOGPATH = "logs/"
	//FORMAT .
	FORMAT = "20060102"
	//LineFeed 换行
	LineFeed = "\r\n"
)

//以天为基准,存日志
var logPath = LOGPATH + time.Now().Format(FORMAT) + "/"

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
				// log.Println(payload)

				//解包负载 payload
				var rsq Log.SendLogReq
				if err := proto.Unmarshal(payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {

					log.Println("日志内容---------------------")
					log.Println("TimeAt: ", rsq.TimeAt)
					log.Println("Username: ", rsq.Username)
					log.Println("DeviceID: ", rsq.DeviceID)
					log.Println("Content: ", rsq.Content)
					t := int64(rsq.TimeAt) / 1e6
					tm := time.Unix(t, 0)
					var logDate string = tm.Format("2006-01-02 15:04:05")
					WriteLog(logDate, rsq.Username, rsq.DeviceID, rsq.Content)

				}

			}
		}

	},
}

//WriteLog return error
func WriteLog(logDate, username, deviceID, msg string) error {
	if !IsExist(logPath) {
		return CreateDir(logPath)
	}
	var (
		err error
		f   *os.File
	)

	f, err = os.OpenFile(logPath+username+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	_, err = io.WriteString(f, ""+logDate+" ["+username+"] "+msg+LineFeed)

	defer f.Close()
	return err
}

//CreateDir  文件夹创建
func CreateDir(logPath string) error {
	err := os.MkdirAll(logPath, os.ModePerm)
	if err != nil {
		return err
	}
	os.Chmod(logPath, os.ModePerm)
	return nil
}

//IsExist  判断文件夹/文件是否存在  存在返回 true
func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func init() {
	rootCmd.AddCommand(sdklogsubCmd)

	sdklogsubCmd.PersistentFlags().StringP("username", "u", "id1", "register username ")
}
