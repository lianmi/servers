/*
用户模块
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("user called")
	},
}

func init() {
	rootCmd.AddCommand(userCmd)

}
