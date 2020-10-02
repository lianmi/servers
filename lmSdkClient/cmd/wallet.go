/*

 */
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// walletCmd represents the wallet command
var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("wallet called")
	},
}

func init() {
	rootCmd.AddCommand(walletCmd)

}
