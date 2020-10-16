/*
操作通用商品
*/
package cmd

import (
	"fmt"

	"github.com/lianmi/servers/lmSdkClient/business/order"
	"github.com/spf13/cobra"
)

// MockGeneralProductCmd represents the MockGeneralProduct command
var MockGeneralProductCmd = &cobra.Command{
	Use:   "MockGeneralProduct",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("MockGeneralProduct called")

		if err := order.MockGeneralProduct(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("MockGeneralProduct run ok")
		}

	},
}

func init() {
	//子命令
	orderCmd.AddCommand(MockGeneralProductCmd)

}
