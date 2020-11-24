/*
商品
*/
package cmd

import (
	// "fmt"

	"github.com/spf13/cobra"
)

// productCmd represents the product command
var productCmd = &cobra.Command{
	Use:   "product",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("product called")
	},
}

func init() {
	rootCmd.AddCommand(productCmd)

}
