/*
7-2 商品上架
*/
package cmd

import (
	// "fmt"
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/order"
	"github.com/spf13/cobra"
)

// addproductCmd represents the addproduct command
var addproductCmd = &cobra.Command{
	Use:   "addproduct",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("addproduct called")
		if err := order.AddProduct(); err != nil {
			log.Println(err)
		}

	},
}

func init() {
	//子命令
	productCmd.AddCommand(addproductCmd)
	// addproductCmd.PersistentFlags().StringP("orderID", "o", "", "订单ID")
}
