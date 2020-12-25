/*
10-1 钱包账号注册
*/
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/wallet"
	"github.com/spf13/cobra"
)

// RegisterWalletCmd represents the RegisterWallet command
var RegisterWalletCmd = &cobra.Command{
	Use:   "RegisterWallet",
	Short: "./lmSdkClient wallet RegisterWallet",
	Long:  `A用户利用钱包SDK生成的地址(约定第0号叶子的地址),  例子： ./lmSdkClient wallet RegisterWallet`,
	Run: func(cmd *cobra.Command, args []string) {
		// username, _ := cmd.PersistentFlags().GetString("username")
		// if username == "" {
		// 	log.Println("username missed")
		// }

		err := wallet.RegisterWallet()
		if err != nil {
			log.Println("RegisterWallet failed")
		} else {
			log.Println("RegisterWallet succeed")
		}
	},
}

func init() {
	// 子命令
	walletCmd.AddCommand(RegisterWalletCmd)
	// RegisterWalletCmd.PersistentFlags().StringP("walletAddress", "u", "", "username")

}
