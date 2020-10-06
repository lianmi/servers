/*
10-5 查询账号余额
*/
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/wallet"
	"github.com/spf13/cobra"
)

// BalanceCmd represents the Balance command
var BalanceCmd = &cobra.Command{
	Use:   "Balance",
	Short: "./lmSdkClient wallet Balance",
	Long:  `查询链上账号余额， 包括连米币及以太币 `,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("Balance called")

		err := wallet.Balance()
		if err != nil {
			log.Println("Balance failed")
		} else {
			log.Println("Balance succeed")
		}

	},
}

func init() {
	//子命令
	walletCmd.AddCommand(BalanceCmd)
}
