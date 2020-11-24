/*
7-4 商品下架

*/
package cmd

import (
	// "fmt"
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/order"

	"github.com/spf13/cobra"
)

// soldoutproductCmd represents the soldoutproduct command
var soldoutproductCmd = &cobra.Command{
	Use:   "soldoutproduct",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("soldoutproduct called")
		if err := order.SoldoutProduct(); err != nil {
			log.Println(err)
		}
	},
}

func init() {
	productCmd.AddCommand(soldoutproductCmd)

}
