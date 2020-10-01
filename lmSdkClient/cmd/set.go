/*
设置本地用户 username及deviceID, 保存在redis
*/
package cmd

import (
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/lmSdkClient/common"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "./lmSdkClient set -d 959bb0ae-1c12-4b60-8741-173361ceba8a",
	Long:  `./lmSdkClient set -d 959bb0ae-1c12-4b60-8741-173361ceba8a`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var deviceid string

		deviceid, err = cmd.PersistentFlags().GetString("deviceid")
		if err != nil {
			log.Println("deviceid is empty")
			return
		}

		redisConn, err := redis.Dial("tcp", common.RedisAddr)
		if err != nil {
			log.Fatalln(err)
			return
		}

		defer redisConn.Close()

		_, err = redisConn.Do("SET", "LocalDeviceID", deviceid)
		if err != nil {
			log.Fatalln(err)
			return
		}
		log.Println("set called")
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.PersistentFlags().StringP("deviceid", "d", "", "deviceid")
}
