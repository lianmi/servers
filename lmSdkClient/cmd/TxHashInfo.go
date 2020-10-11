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
		txType, _ := cmd.PersistentFlags().GetInt32("txType")
		txHash, _ := cmd.PersistentFlags().GetString("txHash")
		// txType := int32(3)
		//0xc69374390821f44a5c93fdffc34fb71e0accf98bc365451d2ca6c33fd94d6f0b
		// txHash := "0x02876f65897b431b4f455e5d72944d0a3b8a3ac0280bddf4f1b1f131cdb865bf"
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
	TxHashInfoCmd.PersistentFlags().Int32P("txType", "p", 3, "交易类型枚举")
	TxHashInfoCmd.PersistentFlags().StringP("txHash", "x", "", "交易哈希")
}
