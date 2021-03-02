/*

 */
package cmd

import (
	"fmt"

	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
)

// tailfCmd represents the tailf command
var tailfCmd = &cobra.Command{
	Use:   "tailf",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tailf called")
		file, _ := cmd.PersistentFlags().GetString("file")
		t, err := tail.TailFile(file, tail.Config{Follow: true})
		for line := range t.Lines {
			fmt.Println(line.Text)
		}
		_ = err
	},
}

func init() {
	rootCmd.AddCommand(tailfCmd)

	tailfCmd.PersistentFlags().StringP("file", "f", "", "log file path")
}
