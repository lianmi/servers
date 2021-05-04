/*
从excel表导入网点数据
*/
package cmd

import (
	// "fmt"

	"fmt"
	"log"

	"github.com/360EntSecGroup-Skylar/excelize/v2"

	"github.com/spf13/cobra"
)

// loadexcelCmd represents the loadexcel command
var loadexcelCmd = &cobra.Command{
	Use:   "loadexcel",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("loadexcel called")
		log.Println("导入当前目录的excel文件")
		f, err := excelize.OpenFile("1619884694.xlsx")
		if err != nil {
			fmt.Println(err)
			return
		}
		rows, err := f.GetRows("Sheet1")
		for _, row := range rows {
			for _, colCell := range row {
				fmt.Print(colCell, "\t")
			}
			fmt.Println()
		}

	},
}

func init() {
	rootCmd.AddCommand(loadexcelCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loadexcelCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loadexcelCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
