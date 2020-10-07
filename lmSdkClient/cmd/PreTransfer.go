/*
10-3 发起转账
*/
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/wallet"
	"github.com/spf13/cobra"
)

// PreTransferCmd represents the PreTransfer command
var PreTransferCmd = &cobra.Command{
	Use:   "PreTransfer",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("PreTransfer called")
		
		orderID, _ := cmd.PersistentFlags().GetString("orderID")
		targetUserName, _ := cmd.PersistentFlags().GetString("targetUserName")
		amount, _ := cmd.PersistentFlags().GetFloat64("amount")

		err := wallet.PreTransfer(orderID, targetUserName, amount)
		if err != nil {
			log.Println("Balance failed")
		} else {
			log.Println("Balance succeed")
		}

	},
}

func init() {
	//子命令
	walletCmd.AddCommand(PreTransferCmd)
	PreTransferCmd.PersistentFlags().StringP("orderID", "o", "", "订单ID")
	PreTransferCmd.PersistentFlags().StringP("targetUserName", "t", "", "收款方的用户账号, like: 0x---------")
	PreTransferCmd.PersistentFlags().Float64P("amount", "a", 0.00, "金额(人民币格式), like: 4.05")

}
