/*
10-10 同步充值历史
*/
package cmd

import (
	"log"

	"github.com/lianmi/servers/lmSdkClient/business/wallet"
	"github.com/spf13/cobra"
)

// SyncDepositHistoryPageCmd represents the SyncDepositHistoryPage command
var SyncDepositHistoryPageCmd = &cobra.Command{
	Use:   "SyncDepositHistoryPage",
	Short: "./lmSdkClient wallet SyncDepositHistoryPage",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("SyncDepositHistoryPage called")
		depositRecharge, _ := cmd.PersistentFlags().GetInt32("depositRecharge")
		startAt, _ := cmd.PersistentFlags().GetInt64("startAt")
		endAt, _ := cmd.PersistentFlags().GetInt64("endAt")
		page, _ := cmd.PersistentFlags().GetInt32("page")
		pageSize, _ := cmd.PersistentFlags().GetInt32("pageSize")

		err := wallet.DoSyncDepositHistoryPage(depositRecharge, startAt, endAt, page, pageSize)
		if err != nil {
			log.Println("DoSyncDepositHistoryPage failed")
		} else {
			log.Println("DoSyncDepositHistoryPage succeed")
		}

	},
}

func init() {
	//子命令
	walletCmd.AddCommand(SyncDepositHistoryPageCmd)
	RegisterWalletCmd.PersistentFlags().Int32P("depositRecharge", "d", 0, "充值金额枚举, 0-不限")
	RegisterWalletCmd.PersistentFlags().Int64P("startAt", "s", 0, "开始时间")
	RegisterWalletCmd.PersistentFlags().Int64P("endAt", "e", 0, "结束时间")
	RegisterWalletCmd.PersistentFlags().Int32P("page", "p", 1, "页数,第几页")
	RegisterWalletCmd.PersistentFlags().Int32P("pageSize", "z", 100, "每页记录数量")

}
