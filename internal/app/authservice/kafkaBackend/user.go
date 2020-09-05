/*
本文件是处理业务号是用户模块，分别有
1-1 获取用户资料 GetUsers
1-2 修改用户资料 UpdateUserProfile
1-3 同步用户资料事件 SyncUserProfileEvent
1-4 同步其它端修改的用户资料 SyncUpdateProfileEvent 未完成
1-5 打标签 MarkTag
1-6 同步其它端标签更改事件 SyncMarkTagEvent
*/
package kafkaBackend

import (
	// "encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	// "strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
)

/*
1. 先从redis里读取 哈希表 userData:{username} 里的元素，如果无法读取，则直接从MySQL里读取
2. 注意，更新资料后，也需要同步更新 哈希表 userData:{username}
哈希表 userData:{username} 的元素就是User的各个字段
*/
func (kc *KafkaClient) HandleGetUsers(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte

	rsp := &User.GetUsersResp{
		Users: make([]*User.User, 0),
	}

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
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {

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
					kc.logger.Error("MySQL里读取错误, 可能记录不存在", zap.Error(err))
					// errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					// errorMsg = fmt.Sprintf("Get user error[username=%s]", username)
					// goto COMPLETE
					continue
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

			rsp.Users = append(rsp.Users, user)

		}

		kc.logger.Info("GetUsers Succeed",
			zap.String("Username:", username))

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ = proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
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

/*
1-2 更新用户资料
将触发 1-4 同步其它端修改的用户资料

*/
func (kc *KafkaClient) HandleUpdateUserProfile(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	rsp := &User.UpdateProfileRsp{}
	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleUpdateUserProfile start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("UpdateUserProfile",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	var req User.UpdateUserProfileReq
	if err := proto.Unmarshal(body, &req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		//查询出需要修改的用户
		pUser := new(models.User)
		if err = kc.db.Model(pUser).Where("username = ?", username).First(pUser).Error; err != nil {
			kc.logger.Error("Query user Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query user Error: %s", err.Error())
			goto COMPLETE
		}

		if nick, ok := req.Fields[1]; ok {
			//修改呢称
			pUser.Nick = nick
		}

		if gender, ok := req.Fields[2]; ok {
			//修改 性别
			pUser.Gender, _ = strconv.Atoi(gender)
		}

		if avatar, ok := req.Fields[3]; ok {
			//修改 头像
			pUser.Avatar = avatar
		}

		if label, ok := req.Fields[4]; ok {
			//修改 签名
			pUser.Label = label
		}

		if email, ok := req.Fields[5]; ok {
			//修改 Email
			pUser.Email = email
		}

		if extend, ok := req.Fields[6]; ok {
			//修改 Extend
			pUser.Extend = extend
		}

		if allowType, ok := req.Fields[7]; ok {
			//修改 AllowType
			pUser.AllowType, _ = strconv.Atoi(allowType)
		}

		if province, ok := req.Fields[8]; ok {
			pUser.Province = province
		}

		if city, ok := req.Fields[9]; ok {
			pUser.City = city
		}

		if county, ok := req.Fields[10]; ok {
			pUser.County = county
		}

		if street, ok := req.Fields[10]; ok {
			pUser.Street = street
		}

		if address, ok := req.Fields[10]; ok {
			pUser.Address = address
		}

		if branches_name, ok := req.Fields[11]; ok {
			pUser.Branchesname = branches_name
		}

		if legal_person, ok := req.Fields[11]; ok {
			pUser.LegalPerson = legal_person
		}

		if legal_identity_card, ok := req.Fields[11]; ok {
			pUser.LegalPerson = legal_identity_card
		}

		pUser.UpdatedAt = time.Now().UnixNano() / 1e6

		if err := kc.SaveUser(pUser); err != nil {
			kc.logger.Error("更新用户信息失败", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Update user error[username=%d]", username)
			goto COMPLETE
		}

		//修改redis里的userData:{username}哈希表，以便GetUsers的时候可以获取最新的数据

		userKey := fmt.Sprintf("userData:%s", username)
		userData := new(models.User)

		if err = kc.db.Model(userData).Where("username = ?", username).First(userData).Error; err != nil {
			kc.logger.Error("MySQL里读取错误", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query user error[username=%s]", username)
			goto COMPLETE
		}
		if _, err := redisConn.Do("HMSET", redis.Args{}.Add(userKey).AddFlat(userData)...); err != nil {
			kc.logger.Error("错误：HMSET", zap.Error(err))
		} else {
			kc.logger.Debug("刷新Redis的用户数据成功", zap.String("username", username))
		}

		//更新redis的sync:{用户账号} myInfoAt 时间戳
		redisConn.Do("HSET",
			fmt.Sprintf("sync:%s", username),
			"myInfoAt",
			time.Now().UnixNano()/1e6)

		//更新redis的sync:{用户账号} friendUsersAt 时间戳 让所有与此用户为好友的用户触发更新用户资料
		//从redis里读取username的好友列表
		friends, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:1", username), "-inf", "+inf"))
		for _, friendUsername := range friends {

			redisConn.Do("HSET",
				fmt.Sprintf("sync:%s", friendUsername),
				"friendUsersAt",
				time.Now().UnixNano()/1e6)
		}

		rsp = &User.UpdateProfileRsp{
			TimeTag: uint64(time.Now().UnixNano() / 1e6),
		}

		kc.logger.Info("UpdateUserProfile Succeed",
			zap.String("Username:", username))

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ := proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("UpdateProfileRsp message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send UpdateProfileRsp message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
1-4 同步其它端修改的用户资料
*/
func (kc *KafkaClient) HandleSyncUpdateProfileEvent(sourceDeviceID string, userData *models.User) error {
	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//构造SyncUpdateProfileEventRsp
	username := userData.Username

	rsp := &User.SyncUpdateProfileEventRsp{
		Fields: make(map[int32]string),
	}

	rsp.Fields[1] = userData.Nick
	rsp.Fields[2] = User.Gender_name[int32(userData.GetGender())]
	rsp.Fields[3] = userData.Avatar
	rsp.Fields[4] = userData.Label
	rsp.Fields[5] = userData.Email
	rsp.Fields[6] = userData.Extend
	rsp.Fields[7] = User.AllowType_name[int32(userData.GetAllowType())]
	rsp.Fields[8] = userData.Province
	rsp.Fields[9] = userData.City
	rsp.Fields[10] = userData.County
	rsp.Fields[11] = userData.Street
	rsp.Fields[12] = userData.Address
	rsp.Fields[13] = userData.Branchesname
	rsp.Fields[14] = userData.LegalPerson
	rsp.Fields[15] = userData.LegalIdentityCard

	data, _ := proto.Marshal(rsp)

	//向其它端响应SyncUpdateProfileEvent事件
	deviceListKey := fmt.Sprintf("devices:%s", username)
	deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
	//查询出当前在线所有主从设备
	for _, eDeviceID := range deviceIDSliceNew {

		//如果设备id是当前操作的，则不发送此事件
		if sourceDeviceID == eDeviceID {
			continue
		}

		targetMsg := &models.Message{}
		curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
		curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
		kc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

		targetMsg.UpdateID()
		//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
		targetMsg.BuildRouter("Auth", "", "Auth.Frontend")

		targetMsg.SetJwtToken(curJwtToken)
		targetMsg.SetUserName(username)
		targetMsg.SetDeviceID(eDeviceID)
		// kickMsg.SetTaskID(uint32(taskId))
		targetMsg.SetBusinessTypeName("User")
		targetMsg.SetBusinessType(uint32(1))
		targetMsg.SetBusinessSubType(uint32(4)) //SyncUpdateProfileEvent = 4

		targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6)

		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

		targetMsg.SetCode(200) //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Auth.Frontend"
		if err := kc.Produce(topic, targetMsg); err == nil {
			kc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			kc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
		}

		kc.logger.Info("Sync myInfoAt Succeed",
			zap.String("Username:", username),
			zap.String("DeviceID:", curDeviceKey),
			zap.Int64("Now", time.Now().UnixNano()/1e6))

	}

	return nil

}

/*
1-5 打标签 MarkTag
*/
func (kc *KafkaClient) HandleMarkTag(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //当前用户账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID() //当前设备id

	kc.logger.Info("HandleMarkTag start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("MarkTag",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	var req User.MarkTagReq
	if err := proto.Unmarshal(body, &req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = "Protobuf Unmarshal Error"
		goto COMPLETE
	} else {
		//修改的用户
		pUser := new(models.User)
		account := req.GetUsername()

		if account == username {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Can't mark yourself.[username=%s]", account)
			goto COMPLETE
		}

		//判断account是否为好友
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Friend:%s:1", username), account); err == nil {
			if reply == nil {
				//A好友列表中没有B
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Can't mark none friend user.[username=%s]", account)
				goto COMPLETE
			}
		}

		if err = kc.db.Model(pUser).Where("username = ?", account).First(pUser).Error; err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query user error[username=%s]", account)
			goto COMPLETE
		}

		if req.GetIsAdd() { //增加
			pTag := new(models.Tag)
			pTag.UserID = pUser.ID
			pTag.UpdatedAt = time.Now().UnixNano() / 1e6
			pTag.TagType = int(req.GetType())

			// //如果已经存在，则先删除，确保不会重复增加
			// where := models.Tag{UserID: pUser.ID, TagType: int(req.GetType())}
			// db := kc.db.Where(&where).Delete(models.Tag{})
			// err = db.Error
			// if err != nil {
			// 	kc.logger.Error("删除Tag出错", zap.Error(err))
			// 	errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			// 	errorMsg = fmt.Sprintf("删除Tag出错[userID=%d]", pUser.ID)
			// 	goto COMPLETE
			// }
			// count := db.RowsAffected
			// kc.logger.Debug("删除Tag成功", zap.Int64("count", count))

			//保存标签MarkTag
			if err := kc.SaveTag(pTag); err != nil {
				kc.logger.Error("MarkTag增加失败", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "MarkTag error"
				goto COMPLETE
			}

			kc.logger.Debug("增加标签成功")

			//Redis缓存黑名单
			if req.GetType() == User.MarkTagType_Mtt_Blocked {
				if _, err = redisConn.Do("ZADD", fmt.Sprintf("BlackList:%s", username), time.Now().UnixNano()/1e6, req.GetUsername()); err != nil {
					kc.logger.Error("ZADD Error", zap.Error(err))
				}
				//在自己的关注有序列表里移除此用户
				if _, err = redisConn.Do("ZREM", fmt.Sprintf("BeWatching:%s", username), req.GetUsername()); err != nil {
					kc.logger.Error("ZREM Error", zap.Error(err))
				}
				//增加用户的取消关注有序列表
				if _, err = redisConn.Do("ZADD", fmt.Sprintf("CancelWatching:%s", username), time.Now().UnixNano()/1e6, req.GetUsername()); err != nil {
					kc.logger.Error("ZADD Error", zap.Error(err))
				}
				//更新redis的sync:{用户账号} watchAt 时间戳
				redisConn.Do("HSET",
					fmt.Sprintf("sync:%s", username),
					"watchAt",
					time.Now().UnixNano()/1e6)
			}

		} else { //删除
			where := models.Tag{UserID: pUser.ID, TagType: int(req.GetType())}
			db := kc.db.Where(&where).Delete(models.Tag{})
			err = db.Error
			if err != nil {
				kc.logger.Error("删除实体出错", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("删除实体出错[userID=%d]", pUser.ID)
				goto COMPLETE
			}
			count := db.RowsAffected

			//删除黑名单，Redis缓存
			if req.GetType() == User.MarkTagType_Mtt_Blocked {
				if _, err = redisConn.Do("ZREM", fmt.Sprintf("BlackList:%s", username), req.GetUsername()); err != nil {
					kc.logger.Error("ZADD Error", zap.Error(err))
				}
				//在自己的关注有序列表里重新加回此用户
				if _, err = redisConn.Do("ZADD", fmt.Sprintf("BeWatching:%s", username), time.Now().UnixNano()/1e6, req.GetUsername()); err != nil {
					kc.logger.Error("ZADD Error", zap.Error(err))
				}
				//更新redis的sync:{用户账号} watchAt 时间戳
				redisConn.Do("HSET",
					fmt.Sprintf("sync:%s", username),
					"watchAt",
					time.Now().UnixNano()/1e6)
			}

			kc.logger.Debug("删除标签成功", zap.Int64("count", count))
		}

		//将标签的变化广播给当前用户的其他端
		{
			deviceListKey := fmt.Sprintf("devices:%s", username)
			deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
			//查询出当前在线所有主从设备
			for _, eDeviceID := range deviceIDSliceNew {

				//自己不发, 其他端才发
				if deviceID == eDeviceID {
					continue
				}

				targetMsg := &models.Message{}
				curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
				curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
				kc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

				targetMsg.UpdateID()
				//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
				targetMsg.BuildRouter("Auth", "", "Auth.Frontend")

				targetMsg.SetJwtToken(curJwtToken)
				targetMsg.SetUserName(username)
				targetMsg.SetDeviceID(eDeviceID)
				// kickMsg.SetTaskID(uint32(taskId))
				targetMsg.SetBusinessTypeName("User")
				targetMsg.SetBusinessType(uint32(2))
				targetMsg.SetBusinessSubType(uint32(3)) //MultiLoginEvent = 3

				targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6)

				//构造负载数据
				resp := &User.SyncMarkTagEventRsp{
					Account: account,
					Type:    req.GetType(),
				}

				targetData, _ := proto.Marshal(resp)
				targetMsg.FillBody(targetData) //网络包的body，承载真正的业务数据

				targetMsg.SetCode(200) //成功的状态码
				//构建数据完成，向dispatcher发送
				topic := "Auth.Frontend"
				if err := kc.Produce(topic, targetMsg); err == nil {
					kc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
				} else {
					kc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
				}
			}
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//只需返回200
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("MarkTag message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send MarkTag message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}
