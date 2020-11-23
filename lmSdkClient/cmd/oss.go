/*
阿里云
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ossCmd represents the oss command
var ossCmd = &cobra.Command{
	Use:   "oss",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("oss called")
	},
}

func init() {
	rootCmd.AddCommand(ossCmd)
}
