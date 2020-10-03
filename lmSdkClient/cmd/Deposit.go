/*
10-2 充值
*/
package cmd

import (
	// "fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/lianmi/servers/lmSdkClient/business/wallet"
)

// DepositCmd represents the Deposit command
var DepositCmd = &cobra.Command{
	Use:   "Deposit",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("Deposit called")
		rechargeAmount, _ := cmd.PersistentFlags().GetFloat64("rechargeAmount")
		err := wallet.Deposit(rechargeAmount)
		if err != nil {
			log.Println("RegisterWallet failed")
		} else {
			log.Println("RegisterWallet succeed")
		}
	},
}

func init() {
	// 子命令
	walletCmd.AddCommand(DepositCmd)
	DepositCmd.PersistentFlags().Float64P("rechargeAmount", "r", 0, "recharge amount, like: 100")

}
