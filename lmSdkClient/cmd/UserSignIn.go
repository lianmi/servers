/*
用户签到
*/
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/wallet"
	"github.com/lianmi/servers/util/dateutil"
	"github.com/spf13/cobra"
)

// UserSignInCmd represents the UserSignIn command
var UserSignInCmd = &cobra.Command{
	Use:   "UserSignIn",
	Short: "./lmSdkClient wallet UserSignIn",
	Long:  `用户每天签到，每成功签到2次，送若干1千万wei的以太币`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("UserSignIn called")

		currDate := dateutil.GetDateString()

		log.Println("currDate", currDate)

		err := wallet.UserSignIn()
		if err != nil {
			log.Println("UserSignIn failed")
		} else {
			log.Println("UserSignIn succeed")
		}
	},
}

func init() {
	//子命令
	walletCmd.AddCommand(UserSignInCmd)
}
