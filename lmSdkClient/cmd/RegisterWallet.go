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
	Short: "用户利用钱包SDK生成的地址(约定第0号叶子的地址)",
	Long:  `A用户利用钱包SDK生成的地址(约定第0号叶子的地址)`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("RegisterWallet called")
		walletAddress, _ := cmd.PersistentFlags().GetString("walletAddress")

		err := wallet.RegisterWallet(walletAddress)
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
	loginCmd.PersistentFlags().StringP("walletAddress", "w", "", "your walletAddress, like: 0x---------")

}
