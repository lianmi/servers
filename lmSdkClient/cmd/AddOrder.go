/*
订单相关
*/
package cmd

import (
	// "fmt"
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/order"
	"github.com/spf13/cobra"
)

// AddOrderCmd represents the AddOrder command
var AddOrderCmd = &cobra.Command{
	Use:   "AddOrder",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("AddOrder called")
		orderID, _ := cmd.PersistentFlags().GetString("orderID")
		productID, _ := cmd.PersistentFlags().GetString("productID")
		if err := order.AddOrder(orderID, productID); err != nil {
			log.Println(err)
		}

	},
}

func init() {
	//子命令
	orderCmd.AddCommand(AddOrderCmd)
	AddOrderCmd.PersistentFlags().StringP("orderID", "o", "", "订单ID")
	AddOrderCmd.PersistentFlags().StringP("productID", "p", "", "商品ID")

}
