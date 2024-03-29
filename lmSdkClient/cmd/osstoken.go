/*
获取oss临时token
*/
package cmd

import (
	"github.com/lianmi/servers/lmSdkClient/business/chat"
	"github.com/spf13/cobra"
	"log"
)

// osstokenCmd represents the osstoken command
var osstokenCmd = &cobra.Command{
	Use:   "osstoken",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		isPrivate, _ := cmd.PersistentFlags().GetBool("isprivate")
		err := chat.GetOssToken(isPrivate)

		if err != nil {
			log.Println("GetOssToken error:", err)
			return
		}

	},
}

func init() {
	//子命令
	ossCmd.AddCommand(osstokenCmd)
	osstokenCmd.PersistentFlags().BoolP("isprivate", "p", false, "用户私有")

}
