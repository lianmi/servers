/*
商户及订单模块 对应文档的第九章
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// orderCmd represents the order command
var orderCmd = &cobra.Command{
	Use:   "order",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("order called")
	},
}

func init() {
	rootCmd.AddCommand(orderCmd)

}
