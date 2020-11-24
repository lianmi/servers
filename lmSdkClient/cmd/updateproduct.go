/*

*/
package cmd

import (
	// "fmt"
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/order"
	"github.com/spf13/cobra"
)

// updateproductCmd represents the updateproduct command
var updateproductCmd = &cobra.Command{
	Use:   "updateproduct",
	Short: "",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("updateproduct called")

		if err := order.UpdateProduct(); err != nil {
			log.Println(err)
		}
	},
}

func init() {
	productCmd.AddCommand(updateproductCmd)
}
