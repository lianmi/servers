package cmd

import (
	// "fmt"
	"github.com/lianmi/servers/lmSdkClient/business/auth"
	"github.com/lianmi/servers/lmSdkClient/common"
	"github.com/lianmi/servers/util/array"
	"github.com/spf13/cobra"
	"log"
)

// smscodeCmd represents the smscode command
var smscodeCmd = &cobra.Command{
	Use:   "smscode",
	Short: "传入手机号，获取验证码",
	Long:  `传入手机号，获取验证码`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("传入手机号，获取验证码:=========================================================")

		mobile, err := cmd.PersistentFlags().GetString("mobile")
		if err != nil {
			log.Println("mobile is empty")
			return
		}
		if mobile == "" {
			log.Println("mobile is empty, 例子: ./lmSdkClient auth smscode -m 12702290109")
			return
		}
		log.Println(mobile)
		client, err := auth.NewClient(common.SERVER_URL, "", false)
		if err != nil {
			log.Fatalln("NewClient error:", err)
		}

		authService := client.NewAuthService()

		response, err := authService.SendSms(mobile)

		if err != nil {
			log.Println("SendSms error:", err)
			return
		}
		array.PrintPretty(response.Get("code")) //200
		array.PrintPretty(response.Get("msg"))  //ok
	},
}

func init() {
	//子命令
	authCmd.AddCommand(smscodeCmd)

	smscodeCmd.PersistentFlags().StringP("mobile", "m", "", "your mobile number")
}
