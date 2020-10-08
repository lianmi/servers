/*
10-11 同步提现历史
*/
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/wallet"
	"github.com/spf13/cobra"
)

// SyncWithdrawHistoryPageCmd represents the SyncWithdrawHistoryPage command
var SyncWithdrawHistoryPageCmd = &cobra.Command{
	Use:   "SyncWithdrawHistoryPage",
	Short: "",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("SyncWithdrawHistoryPage called")

		startAt, _ := cmd.PersistentFlags().GetInt64("startAt")
		endAt, _ := cmd.PersistentFlags().GetInt64("endAt")
		page, _ := cmd.PersistentFlags().GetInt32("page")
		pageSize, _ := cmd.PersistentFlags().GetInt32("pageSize")

		err := wallet.DoSyncWithdrawHistoryPage(startAt, endAt, page, pageSize)
		if err != nil {
			log.Println("DoSyncWithdrawHistoryPage failed")
		} else {
			log.Println("DoSyncWithdrawHistoryPage succeed")
		}

	},
}

func init() {
	//子命令 
	walletCmd.AddCommand(SyncWithdrawHistoryPageCmd)
	SyncWithdrawHistoryPageCmd.PersistentFlags().Int64P("startAt", "s", 0, "开始时间")
	SyncWithdrawHistoryPageCmd.PersistentFlags().Int64P("endAt", "e", 0, "结束时间")
	SyncWithdrawHistoryPageCmd.PersistentFlags().Int32P("page", "p", 1, "页数,第几页")
	SyncWithdrawHistoryPageCmd.PersistentFlags().Int32P("pageSize", "z", 100, "每页记录数量")
}
