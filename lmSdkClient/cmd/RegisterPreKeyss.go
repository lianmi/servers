/*
9-1 商户上传订单DH加密公钥
*/
package cmd

import (
	// "fmt"
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/order"
	"github.com/spf13/cobra"
)

// RegisterPreKeyssCmd represents the RegisterPreKeyss command
var RegisterPreKeyssCmd = &cobra.Command{
	Use:   "RegisterPreKeyss",
	Short: "",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("RegisterPreKeyss called")
		if err := order.RegisterPreKeys(); err != nil {
			log.Println(err)
		}
	},
}

func init() {
	orderCmd.AddCommand(RegisterPreKeyssCmd)

}
