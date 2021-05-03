
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// submitstoreCmd represents the submitstore command
var submitstoreCmd = &cobra.Command{
	Use:   "submitstore",
	Short: "",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("submitstore called")
	},
}

func init() {
	rootCmd.AddCommand(submitstoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// submitstoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// submitstoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
