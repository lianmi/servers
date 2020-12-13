/*
 9-2 获取网点OPK公钥及订单ID

 ./lmSdkClient order GetPreKeyOrderID -p ba89b52c-eb97-4ce1-bb66-90b95cabffd1

*/
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/order"
	"github.com/spf13/cobra"
)

// GetPreKeyOrderIDCmd represents the GetPreKeyOrderID command
var GetPreKeyOrderIDCmd = &cobra.Command{
	Use:   "GetPreKeyOrderID",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		productid, _ := cmd.PersistentFlags().GetString("productid")

		if err := order.GetPreKeyOrderID(productid); err != nil {
			log.Println(err)
		}
	},
}

func init() {
	orderCmd.AddCommand(GetPreKeyOrderIDCmd)

	GetPreKeyOrderIDCmd.PersistentFlags().StringP("productid", "p", "", "productid")

}
