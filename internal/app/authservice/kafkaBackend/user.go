package kafkaBackend

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/internal/pkg/models"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

/*
1. 先从redis里读取 哈希表 userData:{username} 里的元素，如果无法读取，则直接从MySQL里读取
2. 注意，更新资料后，也需要同步更新 哈希表 userData:{username}
哈希表 userData:{username} 的元素有：
Nick
Gender
Avatar
Label
Email
Extend
AllowType
UserType
Introductory
Province
City
County
Street
Address
Branchesname
LegalPerson
LegalIdentityCard //法人身份证不读取
*/
func (kc *KafkaClient) HandleGetUsers(msg *models.Message) error {
	var err error
	kc.logger.Info("HandleGetUsers start...", zap.String("DeviceId", msg.GetDeviceID()))

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

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

	// deviceListKey := fmt.Sprintf("devices:%s", username)

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var getUsersReq User.GetUsersReq
	if err := proto.Unmarshal(body, &getUsersReq); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		return err
	}
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
					return err
				}
			}
		} else {
			kc.logger.Debug("尝试从 MySQL里读取")

			if err = kc.db.Model(userData).Where("username = ?", username).First(userData).Error; err != nil {
				kc.logger.Error("MySQL里读取错误", zap.Error(err))
				return errors.Wrapf(err, "Get user error[username=%s]", username)
			}

			//将数据写入redis，以防下次再从MySQL里读取

			if _, err := redisConn.Do("HMSET", redis.Args{}.Add(userKey).AddFlat(userData)...); err != nil {
				kc.logger.Error("错误：HMSET", zap.Error(err))
				return err
			}
		}
		user := &User.User{
			Username:     userData.Username,
			Nick:         userData.Nick,
			Gender:       userData.GetGender(),
			Avatar:       userData.Avatar,
			Label:        userData.Label,
			Introductory: userData.Introductory,
			Province:     userData.Province,
			City:         userData.City,
			County:       userData.County,
			Street:       userData.Street,
			Address:      userData.Address,
			Branchesname: userData.Branchesname,
			LegalPerson:  userData.LegalPerson,
			// LegalIdentityCard:  userData.LegalIdentityCard,
		}

		getUsersResp.Users = append(getUsersResp.Users, user)

	}

	msg.SetCode(200) //状态码

	data, _ := proto.Marshal(getUsersResp)
	rspHex := strings.ToUpper(hex.EncodeToString(data))

	kc.logger.Info("GetUsers Succeed",
		zap.String("Username:", username),
		zap.Int("length", len(data)),
		zap.String("rspHex", rspHex))

	msg.FillBody(data) //网络包的body，承载真正的业务数据

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

func (kc *KafkaClient) HandleUpdateUserProfile(msg *models.Message) error {
	var err error
	kc.logger.Info("HandleUpdateUserProfile start...", zap.String("DeviceId", msg.GetDeviceID()))

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

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
		return err
	}

	//查询出需要修改的用户
	pUser := new(models.User)
	if err = kc.db.Model(pUser).Where("username = ?", username).First(pUser).Error; err != nil {
		return errors.Wrapf(err, "Get user error[username=%s]", username)
	}

	//使用事务同时更新用户数据和角色数据
	tx := kc.GetTransaction()

	if nick, ok := req.Fields[1]; ok {
		//修改呢称
		pUser.Nick = nick
		if err := tx.Save(pUser).Error; err != nil {
			kc.logger.Error("更新用户Nick失败", zap.Error(err))
			tx.Rollback()
			return errors.Wrapf(err, "更新用户Nick失败[nick=%s]", nick)
		}
	}

	if gender, ok := req.Fields[2]; ok {
		//修改 性别
		pUser.Gender, _ = strconv.Atoi(gender)
		if err := tx.Save(pUser).Error; err != nil {
			kc.logger.Error("更新用户Gender失败", zap.Error(err))
			tx.Rollback()
			return errors.Wrapf(err, "更新用户Gender失败[gender=%d]", gender)
		}
	}

	if avatar, ok := req.Fields[3]; ok {
		//修改 头像
		pUser.Avatar = avatar
		if err := tx.Save(pUser).Error; err != nil {
			kc.logger.Error("更新用户Avatar失败", zap.Error(err))
			tx.Rollback()
			return errors.Wrapf(err, "更新用户Avatar失败[avatar=%s]", avatar)
		}
	}

	if label, ok := req.Fields[4]; ok {
		//修改 签名
		pUser.Label = label
		if err := tx.Save(pUser).Error; err != nil {
			kc.logger.Error("更新用户Label失败", zap.Error(err))
			tx.Rollback()
			return errors.Wrapf(err, "更新用户Label失败[label=%s]", label)
		}
	}

	if email, ok := req.Fields[5]; ok {
		//修改 Email
		pUser.Email = email
		if err := tx.Save(pUser).Error; err != nil {
			kc.logger.Error("更新用户Email失败", zap.Error(err))
			tx.Rollback()
			return errors.Wrapf(err, "更新用户Email失败[label=%s]", email)
		}
	}

	if extend, ok := req.Fields[6]; ok {
		//修改 Extend
		pUser.Extend = extend
		if err := tx.Save(pUser).Error; err != nil {
			kc.logger.Error("更新用户Extend失败", zap.Error(err))
			tx.Rollback()
			return errors.Wrapf(err, "更新用户Extend失败[extend=%s]", extend)
		}
	}

	if allowType, ok := req.Fields[7]; ok {
		//修改 AllowType
		pUser.AllowType, _ = strconv.Atoi(allowType)
		if err := tx.Save(pUser).Error; err != nil {
			kc.logger.Error("更新用户AllowType失败", zap.Error(err))
			tx.Rollback()
			return errors.Wrapf(err, "更新用户AllowType失败[allowType=%d]", allowType)
		}
	}

	//修改UpdateAt
	pUser.UpdatedAt = time.Now().Unix()
	if err := tx.Save(pUser).Error; err != nil {
		kc.logger.Error("更新用户UpdatedAt失败", zap.Error(err))
		tx.Rollback()
		return errors.Wrapf(err, "更新用户UpdatedAt失败[UpdatedAt=%d]", pUser.UpdatedAt)
	}

	//提交
	tx.Commit()

	//修改redis里的userData:{username}哈希表，以便GetUsers的时候可以获取最新的数据

	userKey := fmt.Sprintf("userData:%s", username)
	userData := new(models.User)

	if err = kc.db.Model(userData).Where("username = ?", username).First(userData).Error; err != nil {
		kc.logger.Error("MySQL里读取错误", zap.Error(err))
		return errors.Wrapf(err, "Get user error[username=%s]", username)
	}
	if _, err := redisConn.Do("HMSET", redis.Args{}.Add(userKey).AddFlat(userData)...); err != nil {
		kc.logger.Error("错误：HMSET", zap.Error(err))
		return err
	} else {
		kc.logger.Debug("刷新Redis的用户数据成功", zap.String("username", username))
	}

	msg.SetCode(200) //状态码
	rsp := &User.UpdateProfileRsp{
		TimeTag: uint64(time.Now().UnixNano() / 1e6),
	}
	data, _ := proto.Marshal(rsp)
	rspHex := strings.ToUpper(hex.EncodeToString(data))

	kc.logger.Info("UpdateUserProfile Succeed",
		zap.String("Username:", username),
		zap.Int("length", len(data)),
		zap.String("rspHex", rspHex))

	msg.FillBody(data) //网络包的body，承载真正的业务数据

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

func (kc *KafkaClient) HandleMarkTag(msg *models.Message) error {
	var err error
	kc.logger.Info("HandleMarkTag start...", zap.String("DeviceId", msg.GetDeviceID()))

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

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
		return err
	}

	//修改的用户
	pUser := new(models.User)
	account := req.GetUsername()
	if err = kc.db.Model(pUser).Where("username = ?", account).First(pUser).Error; err != nil {
		return errors.Wrapf(err, "Get user error[username=%s]", account)
	}
	userID := pUser.ID

	if req.GetIsAdd() { //增加
		pTag := new(models.Tag)
		pTag.UserID = userID
		pTag.UpdatedAt = time.Now().Unix()
		pTag.TagType = int(req.GetType())

		//如果已经存在，则先删除，确保不会重复增加
		where := models.Tag{UserID: userID, TagType: int(req.GetType())}
		db := kc.db.Where(where).Delete(models.Tag{})
		err = db.Error
		if err != nil {
			kc.logger.Error("删除实体出错", zap.Error(err))
			return errors.Wrapf(err, "删除实体出错[userID=%d]", userID)
		}
		count := db.RowsAffected
		kc.logger.Debug("删除实体成功", zap.Int64("count", count))

		//使用事务同时更新用户数据和角色数据
		tx := kc.GetTransaction()

		if err := tx.Save(pUser).Error; err != nil {
			kc.logger.Error("MarkTag增加失败", zap.Error(err))
			tx.Rollback()
			return errors.Wrapf(err, "MarkTag增加失败")
		}
		kc.logger.Debug("增加标签成功")

		//提交
		tx.Commit()

	} else { //删除
		where := models.Tag{UserID: userID, TagType: int(req.GetType())}
		db := kc.db.Where(where).Delete(models.Tag{})
		err = db.Error
		if err != nil {
			kc.logger.Error("删除实体出错", zap.Error(err))
			return errors.Wrapf(err, "删除实体出错[userID=%d]", userID)
		}
		count := db.RowsAffected
		kc.logger.Debug("删除标签成功", zap.Int64("count", count))
	}

	msg.SetCode(200) //状态码

	kc.logger.Info("MarkTag Succeed",
		zap.String("Username:", username))

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
