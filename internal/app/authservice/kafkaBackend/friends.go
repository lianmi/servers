/*
本文件是处理业务号是好友模块，分别有
3-1 好友请求发起与处理 FriendRequest 未完成
3-2 好友关系变更事件 FriendChangeEvent 未完成
3-3 好友列表同步事件 未完成
3-4 好友资料同步事件 未完成
3-5 移除好友
3-6 刷新好友资料
3-7 主从设备好友资料同步事件
3-8 增量同步好友列表
*/
package kafkaBackend

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
)

/*

 */
func (kc *KafkaClient) HandleFriendRequest(msg *models.Message) error {
	var err error
	var errorMsg string

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleGetUsers start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("GetUsers",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var getUsersReq User.GetUsersReq
	if err := proto.Unmarshal(body, &getUsersReq); err != nil {
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		getUsersResp := &User.GetUsersResp{
			Users: make([]*User.User, 0),
		}

		for _, username := range getUsersReq.GetUsernames() {
			kc.logger.Debug("for .. range ...", zap.String("username", username))
			//先从Redis里读取，不成功再从 MySQL里读取
			userKey := fmt.Sprintf("userData:%s", username)

			userData := new(models.User)

			isExists, _ := redis.Bool(redisConn.Do("EXISTS", userKey))
			if isExists {
				if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
					if err := redis.ScanStruct(result, userData); err != nil {

						kc.logger.Error("错误：ScanStruct", zap.Error(err))
						continue

					}
				}
			} else {
				kc.logger.Debug("尝试从 MySQL里读取")

				if err = kc.db.Model(userData).Where("username = ?", username).First(userData).Error; err != nil {
					kc.logger.Error("MySQL里读取错误", zap.Error(err))
					errorMsg = fmt.Sprintf("Get user error[username=%s]", username)
					goto COMPLETE
				}

				//将数据写入redis，以防下次再从MySQL里读取, 如果错误也不会终止
				if _, err := redisConn.Do("HMSET", redis.Args{}.Add(userKey).AddFlat(userData)...); err != nil {
					kc.logger.Error("错误：HMSET", zap.Error(err))
				}
			}
			user := &User.User{
				Username:          userData.Username,
				Nick:              userData.Nick,
				Gender:            userData.GetGender(),
				Avatar:            userData.Avatar,
				Label:             userData.Label,
				Introductory:      userData.Introductory,
				Province:          userData.Province,
				City:              userData.City,
				County:            userData.County,
				Street:            userData.Street,
				Address:           userData.Address,
				Branchesname:      userData.Branchesname,
				LegalPerson:       userData.LegalPerson,
				LegalIdentityCard: userData.LegalIdentityCard,
			}

			getUsersResp.Users = append(getUsersResp.Users, user)

		}
		data, _ := proto.Marshal(getUsersResp)
		rspHex := strings.ToUpper(hex.EncodeToString(data))

		kc.logger.Info("GetUsers Succeed",
			zap.String("Username:", username),
			zap.Int("length", len(data)),
			zap.String("rspHex", rspHex))

		msg.FillBody(data) //网络包的body，承载真正的业务数据

	}

COMPLETE:
	if err != nil {
		msg.SetCode(400)                  //状态码
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)

	} else {
		msg.SetCode(200) //状态码
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("GetUsersResp message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send GetUsersResp message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}
