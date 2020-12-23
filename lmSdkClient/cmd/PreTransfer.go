/*
10-3 发起转账
向id4转账 1 元
./lmSdkClient wallet PreTransfer -t id4 -a 1.00

购买Vip会员
./lmSdkClient order BuyVipUser  -p ada166df-bb9f-4274-ab8d-e369a68d69ce -I 9.9
./lmSdkClient wallet PreTransfer -t id3 -a 9.9 -o 959bb0ae-1c12-4b60-8741-173361ceba8a

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
	PreTransferCmd.PersistentFlags().Float64P("amount", "a", 0.00, "金额(单位是元，人民币格式), like: 4.05")

}
