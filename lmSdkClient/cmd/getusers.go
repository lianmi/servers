/*
1-1 获取用户资料
*/
package cmd

import (
	"fmt"

	"github.com/lianmi/servers/lmSdkClient/business/user"
	"github.com/spf13/cobra"
)

// getusersCmd represents the getusers command
var getusersCmd = &cobra.Command{
	Use:   "getusers",
	Short: "根据用户ID批量获取用户信息,登录后拉取其他用户资料,添加好友查询好友资料",
	Long:  `根据用户ID批量获取用户信息,登录后拉取其他用户资料,添加好友查询好友资料`,
	Run: func(cmd *cobra.Command, args []string) {
		userNames := make([]string, 0)
		for _, username := range args {
			fmt.Println("username:", username)
			userNames = append(userNames, username)
		}
		user.SendGetUsers(userNames)

	},
}

func init() {
	// 子命令
	userCmd.AddCommand(getusersCmd)

}
