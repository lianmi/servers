/*

 */
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/order"
	"github.com/spf13/cobra"
)

// queryproductsCmd represents the queryproducts command
var queryproductsCmd = &cobra.Command{
	Use:   "queryproducts",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("queryproducts called")

		if err := order.QueryProducts(); err != nil {
			log.Println(err)
		}
	},
}

func init() {
	productCmd.AddCommand(queryproductsCmd)

}
