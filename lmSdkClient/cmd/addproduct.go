/*
7-2 商品上架
用户必须是商户类型才能上架商品
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
	
}
