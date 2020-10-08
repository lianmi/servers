/*
10-9 同步收款历史
*/
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/wallet"

	"github.com/spf13/cobra"
)

// SyncCollectionHistoryPageCmd represents the SyncCollectionHistoryPage command
var SyncCollectionHistoryPageCmd = &cobra.Command{
	Use:   "SyncCollectionHistoryPage",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("SyncCollectionHistoryPage called")
		fromUsername, _ := cmd.PersistentFlags().GetString("fromUsername")
		startAt, _ := cmd.PersistentFlags().GetInt64("startAt")
		endAt, _ := cmd.PersistentFlags().GetInt64("endAt")
		page, _ := cmd.PersistentFlags().GetInt32("page")
		pageSize, _ := cmd.PersistentFlags().GetInt32("pageSize")

		err := wallet.DoSyncCollectionHistoryPage(fromUsername, startAt, endAt, page, pageSize)
		if err != nil {
			log.Println("DoSyncCollectionHistoryPage failed")
		} else {
			log.Println("DoSyncCollectionHistoryPage succeed")
		}

	},
}

func init() {
	//子命令
	walletCmd.AddCommand(SyncCollectionHistoryPageCmd)

	SyncCollectionHistoryPageCmd.PersistentFlags().StringP("fromUsername", "f", "", "发起方")
	SyncCollectionHistoryPageCmd.PersistentFlags().Int64P("startAt", "s", 0, "开始时间")
	SyncCollectionHistoryPageCmd.PersistentFlags().Int64P("endAt", "e", 0, "结束时间")
	SyncCollectionHistoryPageCmd.PersistentFlags().Int32P("page", "p", 1, "页数,第几页")
	SyncCollectionHistoryPageCmd.PersistentFlags().Int32P("pageSize", "z", 100, "每页记录数量")
}
