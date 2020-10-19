/*
6-1 发起同步请求
*/
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/sync"
	"github.com/spf13/cobra"
)

// SyncEventCmd represents the SyncEvent command
var SyncEventCmd = &cobra.Command{
	Use:   "SyncEvent",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("SyncEvent called")
		err := sync.DoSyncEvent()
		if err != nil {
			log.Println("DoSyncEvent failed")
		} else {
			log.Println("DoSyncEvent succeed")
		}

	},
}

func init() {
	//子命令
	syncCmd.AddCommand(SyncEventCmd)
}
