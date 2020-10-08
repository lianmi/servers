/*
10-6 发起提现预审核
*/
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/wallet"
	"github.com/spf13/cobra"
)

// PreWithDrawCmd represents the PreWithDraw command
var PreWithDrawCmd = &cobra.Command{
	Use:   "PreWithDraw",
	Short: "./lmSdkClient wallet PreWithDraw -r 100 -s 123456 -b ChinaBank -c 23423423423423423432432432 -w lishijia",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("PreWithDraw called")
		amount, _ := cmd.PersistentFlags().GetFloat64("amount")
		smscode, _ := cmd.PersistentFlags().GetString("smscode")
		bank, _ := cmd.PersistentFlags().GetString("bank")
		bankCard, _ := cmd.PersistentFlags().GetString("bankCard")
		cardOwner, _ := cmd.PersistentFlags().GetString("cardOwner")

		err := wallet.PreWithDraw(amount, smscode, bank, bankCard, cardOwner)
		if err != nil {
			log.Println("PreWithDraw failed")
		} else {
			log.Println("PreWithDraw succeed")
		}

	},
}

func init() {
	//子命令
	walletCmd.AddCommand(PreWithDrawCmd)
	PreWithDrawCmd.PersistentFlags().Float64P("amount", "r", 0, "人民币格式, like: 100.00")
	PreWithDrawCmd.PersistentFlags().StringP("smscode", "s", "123456", "code received from mobile, like: 123456")
	PreWithDrawCmd.PersistentFlags().StringP("bank", "b", "ChinaBank", "")
	PreWithDrawCmd.PersistentFlags().StringP("bankCard", "c", "62212214553434342332", "")
	PreWithDrawCmd.PersistentFlags().StringP("cardOwner", "w", "lishijia", "")

}
