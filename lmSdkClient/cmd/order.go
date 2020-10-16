/*
商品及订单模块
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
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("order called")
	},
}

func init() {
	rootCmd.AddCommand(orderCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// orderCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// orderCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
