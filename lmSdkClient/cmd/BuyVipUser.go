/*
购买Vip会员
*/
package cmd

import (
	// "fmt"
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/order"
	"github.com/spf13/cobra"
)

// BuyVipUserCmd represents the BuyVipUser command
var BuyVipUserCmd = &cobra.Command{
	Use:   "BuyVipUser",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("BuyVipUser called")
		orderID, _ := cmd.PersistentFlags().GetString("orderID")
		productID, _ := cmd.PersistentFlags().GetString("productID")
		price, _ := cmd.PersistentFlags().GetFloat64("price")
		if err := order.BuyVipUser(price, orderID, productID); err != nil {
			log.Println(err)
		}

	},
}

func init() {
	//子命令
	orderCmd.AddCommand(BuyVipUserCmd)
	BuyVipUserCmd.PersistentFlags().StringP("orderID", "o", "", "订单ID")
	BuyVipUserCmd.PersistentFlags().StringP("productID", "p", "", "商品ID")
	BuyVipUserCmd.PersistentFlags().Float64P("price", "I", 0.0, "会员价格")

}
