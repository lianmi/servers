package kafkaBackend

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	Team "github.com/lianmi/servers/api/proto/team"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)


/*
4-1 创建群组
*/
func (kc *KafkaClient) HandleCreateTeam(msg *models.Message) error {
	var err error
	var errorMsg string
	rsp := &Team.CreateTeamRsp{}
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleCreateTeam start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("CreateTeam",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.CreateTeamReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("CreateTeam body",
			zap.String("群主账号", req.GetOwner()),
			zap.Int("群类型", int(req.GetType())), // Normal(1) - 普通群 Advanced(2) - vip群
			zap.String("群组名称", req.GetName()),
			zap.Int("verifyType", int(req.GetVerifyType())), //如果普通群，只能选4，如果vip群，可以选其它
			zap.Int("inviteMode", int(req.GetInviteMode())), //邀请模式
		)

		teamOwner := req.GetOwner()

		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("userData:%s", teamOwner))); err != nil {
			errorMsg = fmt.Sprintf("Query user error[teamOwner=%s]", teamOwner)
			goto COMPLETE

		} else {
			if !isExists {
				err = errors.Wrapf(err, "Owner is not exists[teamOwner=%s]", teamOwner)
				errorMsg = fmt.Sprintf("Owner is not exists[teamOwner=%s]", teamOwner)
				goto COMPLETE
			}

			//判断群主是否已经注册为网点用户类型
			userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", teamOwner), "UserType"))
			if userType != 2 {
				err = errors.Wrapf(err, "userType is not business type [userType=%d]", userType)
				errorMsg = fmt.Sprintf("serType is not business type [userType=%d]", userType)
				goto COMPLETE
			}

			//写入MySQL数据库

			var newTeamIndex uint64
			if newTeamIndex, err = redis.Uint64(redisConn.Do("INCR", "TeamIndex")); err != nil {
				kc.logger.Error("redisConn GET TeamIndex Error", zap.Error(err))
				errorMsg = fmt.Sprintf("serType is not business type [userType=%d]", userType)
				goto COMPLETE
			}

			pTeam := new(models.Team)
			pTeam.CreatedAt = time.Now().Unix()
			pTeam.TeamID = fmt.Sprintf("team%d", newTeamIndex) //群id， 自动生成
			pTeam.Teamname = req.GetName()
			pTeam.Nick = req.GetName()
			pTeam.Owner = req.GetOwner()
			pTeam.Type = int(req.GetType())
			pTeam.VerifyType = int(req.GetVerifyType())
			pTeam.InviteMode = int(req.GetInviteMode())

			//默认的设置
			pTeam.Status = 1 //Init(1) - 初始状态,审核中 Normal(2) - 正常状态 Blocked(3) - 封禁状态
			pTeam.MemberLimit = 600
			pTeam.MemberNum = 1  //刚刚建群是只有群主
			pTeam.MuteType = 1   //None(1) - 所有人可发言
			pTeam.InviteMode = 1 //邀请模式,初始为1

			if err = kc.SaveCreateTeam(pTeam); err != nil {
				kc.logger.Error("Save CreateTeam Error", zap.Error(err))
				errorMsg = "无法保存到数据库"
				goto COMPLETE
			}

			//这里不写入redis，等管理员审核通过后才写入redis里

			//回包
			rsp.TeamInfo = &Team.TeamInfo{
				TeamId:       pTeam.TeamID,
				Name:         pTeam.Teamname,
				Icon:         "",
				Announcement: "",
				Introduce:    "",
				Owner:        pTeam.Owner,
				Type:         Team.TeamType(pTeam.Type),
				VerifyType:   Team.VerifyType(pTeam.VerifyType),
				MemberLimit:  int32(pTeam.MemberLimit),
				MemberNum:    int32(pTeam.MemberNum),
				Status:       Team.Status(pTeam.Status),
				MuteType:     Team.MuteMode(pTeam.MuteType),
				InviteMode:   Team.InviteMode(pTeam.InviteMode),
				Ex:           pTeam.Extend,
			}

			data, _ = proto.Marshal(rsp)
		}
	}

COMPLETE:
	if err != nil {
		msg.SetCode(400)                  //状态码
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)

	} else {
		msg.SetCode(200) //状态码
		data, _ = proto.Marshal(rsp)
		msg.FillBody(data)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}
