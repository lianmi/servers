/*
10-14查询交易哈希详情
*/
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/wallet"
	"github.com/spf13/cobra"
)

// TxHashInfoCmd represents the TxHashInfo command
var TxHashInfoCmd = &cobra.Command{
	Use:   "TxHashInfo",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("TxHashInfo called")
		// txType, _ := cmd.PersistentFlags().GetInt32("txType")
		// txHash, _ := cmd.PersistentFlags().GetString("txHash")
		txType := int32(3)
		txHash := "0x02876f65897b431b4f455e5d72944d0a3b8a3ac0280bddf4f1b1f131cdb865bf"
		err := wallet.DoTxHashInfo(txType, txHash)
		if err != nil {
			log.Println("DoTxHashInfo failed")
		} else {
			log.Println("DoTxHashInfo succeed")
		}
	},
}

func init() {
	walletCmd.AddCommand(TxHashInfoCmd)
	// TxHashInfoCmd.PersistentFlags().Int32P("txType", "p", 1, "交易类型枚举")
	// TxHashInfoCmd.PersistentFlags().StringP("txHash", "h", "", "交易哈希")
}
