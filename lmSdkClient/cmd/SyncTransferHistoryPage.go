/*
10-12 同步转账历史
*/
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/wallet"
	"github.com/spf13/cobra"
)

// SyncTransferHistoryPageCmd represents the SyncTransferHistoryPage command
var SyncTransferHistoryPageCmd = &cobra.Command{
	Use:   "SyncTransferHistoryPage",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("SyncTransferHistoryPage called")
		startAt, _ := cmd.PersistentFlags().GetInt64("startAt")
		endAt, _ := cmd.PersistentFlags().GetInt64("endAt")
		page, _ := cmd.PersistentFlags().GetInt32("page")
		pageSize, _ := cmd.PersistentFlags().GetInt32("pageSize")

		err := wallet.DoSyncTransferHistoryPage(startAt, endAt, page, pageSize)
		if err != nil {
			log.Println("DoSyncTransferHistoryPage failed")
		} else {
			log.Println("DoSyncTransferHistoryPage succeed")
		}

	},
}

func init() {

	//子命令
	walletCmd.AddCommand(SyncTransferHistoryPageCmd)

	SyncTransferHistoryPageCmd.PersistentFlags().Int64P("startAt", "s", 0, "开始时间")
	SyncTransferHistoryPageCmd.PersistentFlags().Int64P("endAt", "e", 0, "结束时间")
	SyncTransferHistoryPageCmd.PersistentFlags().Int32P("page", "p", 1, "页数,第几页")
	SyncTransferHistoryPageCmd.PersistentFlags().Int32P("pageSize", "z", 100, "每页记录数量")
}
