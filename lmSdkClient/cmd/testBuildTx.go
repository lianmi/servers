/*
测试小吴 的A签 是否一致
*/
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/wallet"
	"github.com/spf13/cobra"
)

// testBuildTxCmd represents the testBuildTx command
var testBuildTxCmd = &cobra.Command{
	Use:   "testBuildTx",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("testBuildTx called")
		err := wallet.TestBuildTx()
		if err != nil {
			log.Println("TestBuildTx failed")
		} else {
			log.Println("TestBuildTx succeed")
		}
	},
}

func init() {
	walletCmd.AddCommand(testBuildTxCmd)

}
