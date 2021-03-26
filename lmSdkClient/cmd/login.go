/*
POST /login
{
    "username" : "id1",
    "password" : "C33367701511B4F6020EC61DED352059",
    "smscode" : "123456",
    "deviceid" : "959bb0ae-1c12-4b60-8741-173361ceba8a",
    "clientype": 5,
    "os": "MacOSX",
    "protocolversion": "2.0",
    "sdkversion" : "3.0",
    "ismaster" : true
}
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/lmSdkClient/business/auth"
	"github.com/lianmi/servers/lmSdkClient/common"
	"github.com/lianmi/servers/util/array"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "./lmSdkClient auth login -u id2 -p C33367701511B4F6020EC61DED352059 -s 123456",
	Long:  `./lmSdkClient auth login -u id2 -p C33367701511B4F6020EC61DED352059 -s 123456`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("登录:=========================================================")

		username, err := cmd.PersistentFlags().GetString("username")
		if err != nil {
			log.Println("username is empty")
			return
		}
		password, err := cmd.PersistentFlags().GetString("password")
		if password == "" {
			log.Println("password is empty")
			return
		}
		smscode, err := cmd.PersistentFlags().GetString("smscode")
		if smscode == "" {
			log.Println("smscode is empty")
			return
		}
		deviceID, err := cmd.PersistentFlags().GetString("deviceid")
		if deviceID == "" {
			log.Println("deviceid is empty")
			return
		}
		userType, err := cmd.PersistentFlags().GetInt("userType")
		if userType == 0 {
			userType = 1
		}
		os, err := cmd.PersistentFlags().GetString("os")
		if os == "" {
			log.Println("os is empty")
			return
		}
		protocolVersion, err := cmd.PersistentFlags().GetString("protocolversion")
		if protocolVersion == "" {
			log.Println("protocolVersion is empty")
			return
		}
		sdkVersion, err := cmd.PersistentFlags().GetString("sdkversion")
		if sdkVersion == "" {
			log.Println("sdkversion is empty")
			return
		}
		isMaster, err := cmd.PersistentFlags().GetBool("ismaster")
		if err != nil {
			log.Println(err)
			return
		}

		log.Println(username)
		log.Println(password)
		log.Println(smscode)
		log.Println(deviceID)
		log.Println(userType)
		log.Println(os)
		log.Println(protocolVersion)
		log.Println(sdkVersion)
		log.Println(isMaster)

		login := &auth.Login{
			Username:        username,
			Password:        password,
			SmsCode:         smscode,
			DeviceID:        deviceID,
			UserType:        userType,
			Os:              os,
			ProtocolVersion: protocolVersion,
			SdkVersion:      sdkVersion,
			IsMaster:        isMaster,
		}

		client, err := auth.NewClient(common.SERVER_URL, "", false)
		if err != nil {
			log.Fatalln("NewClient error:", err)
		}

		authService := client.NewAuthService()

		response, err := authService.Login(login)

		if err != nil {
			log.Println("SendSms error:", err)
			return
		}
		array.PrintPretty(response.Get("code")) //200
		array.PrintPretty(response.Get("msg"))
		array.PrintPretty(response.Get("data")) //jwt

		redisConn, err := redis.Dial("tcp", common.RedisAddr)
		if err != nil {
			log.Fatalln(err)
			return
		}

		defer redisConn.Close()

		code, _ := response.Get("code").Int()
		if code == 200 {
			_, err := redisConn.Do("SET", "LocalUserName", username)
			if err != nil {
				log.Fatalln(err)
				return
			}
			_, err = redisConn.Do("SET", "LocalDeviceID", deviceID)
			if err != nil {
				log.Fatalln(err)
				return
			}

			jwtToken, _ := response.Get("data").Get("jwt_token").String()
			key := fmt.Sprintf("jwtToken:%s", username)
			_, err = redisConn.Do("SET", key, jwtToken)
			if err != nil {
				log.Fatalln(err)
				return
			}
			log.Println("Login succesd")
		} else {
			log.Println("Login failure, code != 200")
		}
	},
}

func init() {
	//子命令
	authCmd.AddCommand(loginCmd)

	loginCmd.PersistentFlags().StringP("username", "u", "id1", "your username, like: id1")
	loginCmd.PersistentFlags().StringP("password", "p", "C33367701511B4F6020EC61DED352059", "your password, like: C33367701511B4F6020EC61DED352059")
	loginCmd.PersistentFlags().StringP("smscode", "s", "123456", "code received from mobile, like: 123456")
	loginCmd.PersistentFlags().StringP("deviceid", "d", "959bb0ae-1c12-4b60-8741-173361ceba8a", "deviceid, like: 959bb0ae-1c12-4b60-8741-173361ceba8a")
	loginCmd.PersistentFlags().IntP("userType", "t", 1, "userType, like: 1")
	loginCmd.PersistentFlags().StringP("os", "o", "MacOSX", "os, like: MacOSX")
	loginCmd.PersistentFlags().StringP("protocolversion", "v", "2.0", "protocolversion, like: 2.0")
	loginCmd.PersistentFlags().StringP("sdkversion", "k", "3.0", "sdkversion, like: 3.0")
	loginCmd.PersistentFlags().BoolP("ismaster", "i", true, "ismaster, like: true/false")
}
