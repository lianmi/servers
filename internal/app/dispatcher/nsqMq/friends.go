/*
本文件是处理业务号是好友模块，分别有
3-1 好友请求发起与处理 FriendRequest
<已取消> 3-2 好友关系变更事件 FriendChangeEvent</已取消>
3-3 好友列表同步事件 SyncFriendsEvent
3-4 好友资料同步事件 SyncFriendUsersEvent
3-5 移除好友 DeleteFriend
3-6 修改好友资料 UpdateFriend
3-7 主从设备好友资料同步事件 SyncUpdateFriendEvent
3-8 增量同步好友列表 未完成

用户与商户互为好友或关注的有序集合是Business:%s
*/
package nsqMq

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"

	// User "github.com/lianmi/servers/api/proto/user"
	Friends "github.com/lianmi/servers/api/proto/friends"
	LMCError "github.com/lianmi/servers/internal/pkg/lmcerror"

	Msg "github.com/lianmi/servers/api/proto/msg"
	User "github.com/lianmi/servers/api/proto/user"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	uuid "github.com/satori/go.uuid"

	"go.uber.org/zap"
)

/*
3-1 好友请求发起与处理
预审核好友列表： Friend:{username}:0
当前好友列表： Friend:{username}:1
移除的好友列表 ： Friend:{username}:2

注意：
1. Alice加Bob, 先判断Bob是否已经注册，不用判断Alice是否注册，因为在 dispatcher已经做了这个工作。
2. Bob是否允许加好友, 如果Bob拒绝任何人添加好友，就直接返回给Alice。
   如果Bob运行任何人加好友，就直接互加成功, 并且通过5-2向双方推送加好友成功。
   如果Bob的加好友设定是需要confirm，则需要发系统消息5-2给Bob， 让Bob同意或拒绝。
3. 服务端利用redis的哈希表，保存Alice加Bob的状态，当Bob同意或拒绝后，才进行入库及更新Alice的好友表
4. 要考虑到多端的环境，交互的动作可以在任一端进行，结果需要同步给其他端
5. 以有序集合存储所发生的系统通知， 当已经有了最终结果后，这个有序集合就会只保留最后一个结果，
   如果长时间离线再重新上线的其他端，会收到最后一个结果，而不会重现整个交互流程。
*/
func (nc *NsqClient) HandleFriendRequest(msg *models.Message) error {
	var err error
	errorCode := 200

	var isExists bool

	rsp := &Friends.FriendRequestRsp{}
	var data []byte

	var isAhaveB, isBhaveA bool //A好友列表里有B， B好友列表里有A
	var allowType, userType int

	var newSeq uint64

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleFriendRequest start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("FriendRequest",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Friends.FriendRequestReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		nc.logger.Debug("FriendRequest body",
			zap.String("Username", req.GetUsername()),
			zap.String("Ps", req.GetPs()),         // 附言
			zap.String("Source", req.GetSource()), //来源
			zap.Int("Type", int(req.GetType())))

		psSource := &Friends.PsSource{
			Ps:     req.GetPs(),
			Source: req.GetSource(),
		}
		psSourceData, _ := proto.Marshal(psSource)

		//查出 targetUser 有效性，是否已经是好友，好友增加的设置等信息
		targetKey := fmt.Sprintf("userData:%s", req.GetUsername())

		//检测目标用户是否注册及获取他的添加好友的设定
		if isExists, err = redis.Bool(redisConn.Do("EXISTS", targetKey)); err != nil {
			//redis出错
			err = errors.Wrapf(err, "user not exists[username=%s]", req.GetUsername())
			errorCode = LMCError.RedisError
			goto COMPLETE
		}
		if !isExists {
			//B不存在
			err = errors.Wrapf(err, "user not exists[username=%s]", req.GetUsername())
			errorCode = LMCError.UserNotExistsError
			goto COMPLETE
		}

		//Bob的加好友设定
		allowType, _ = redis.Int(redisConn.Do("HGET", targetKey, "AllowType"))
		//Bob的用户类型
		userType, _ = redis.Int(redisConn.Do("HGET", targetKey, "UserType"))

		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Friend:%s:1", username), req.GetUsername()); err == nil {
			if reply == nil {
				//A好友列表中没有B
				isAhaveB = false
			} else {
				isAhaveB = true
			}

		}
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Friend:%s:1", req.GetUsername()), username); err == nil {
			if reply == nil {
				//B好友列表中没有A
				isBhaveA = false
			} else {
				isBhaveA = true
			}

		}
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("BlackList:%s:1", req.GetUsername()), username); err == nil {
			if reply != nil {
				nc.logger.Warn("用户已被对方拉黑， 不能加好友", zap.String("Username", req.GetUsername()))
				errorCode = LMCError.IsBlackUserError
				goto COMPLETE
			}
		}
		//如果已经互为好友，就直接回复
		if isAhaveB && isBhaveA {
			nc.logger.Debug(fmt.Sprintf("已经互为好友，就直接回复, userA: %s, userB: %s", username, req.GetUsername()))

			err = nil
			rsp.Username = username
			rsp.Status = Friends.OpStatusType_Ost_ApplySucceed
			goto COMPLETE
		}

		//根据操作类型OptType进行逻辑处理
		switch Friends.OptType(req.GetType()) {
		case Friends.OptType_Fr_ApplyFriend: //发起加好友验证
			{
				userA := username          //发起方
				userB := req.GetUsername() //被加方
				nc.logger.Debug(fmt.Sprintf("OptType_Fr_ApplyFriend, userA: %s, userB: %s", userA, userB))

				if allowType == LMCommon.DenyAny { //拒绝任何人添加好友
					nc.logger.Debug("拒绝任何人添加好友")
					rsp.Status = Friends.OpStatusType_Ost_RejectFriendApply
					body := &Msg.MessageNotificationBody{
						Type:           Msg.MessageNotificationType_MNT_RejectFriendApply, //拒绝好友请求
						HandledAccount: userB,
						HandledMsg:     "拒绝任何人添加好友",
						Status:         Msg.MessageStatus_MOS_Done,
						Data:           nil,
						To:             userA,
					}
					bodyData, _ := proto.Marshal(body)
					eRsp := &Msg.RecvMsgEventRsp{
						Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
						Type:         Msg.MessageType_MsgType_Notification, //通知类型
						Body:         bodyData,
						From:         username,    //谁发的
						FromDeviceId: deviceID,    //哪个设备发的
						Recv:         userA,       //接收方, 根据场景判断to是个人还是群
						ServerMsgId:  msg.GetID(), //服务器分配的消息ID
						WorkflowID:   "",
						Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
						Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
						Time:         uint64(time.Now().UnixNano() / 1e6),
					}

					go func() {
						time.Sleep(100 * time.Millisecond)
						nc.logger.Debug("5-2, userB拒绝任何人添加好友",
							zap.String("userB", userB),
							zap.String("to", userA),
						)
						nc.BroadcastSystemMsgToAllDevices(eRsp, userA)
					}()

				} else if allowType == LMCommon.AllowAny { //允许任何人加为好友

					nc.logger.Debug("允许任何人加为好友")

					rsp.Username = username
					rsp.Status = Friends.OpStatusType_Ost_ApplySucceed

					//在A的预审核好友列表里删除B ZREM
					if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:0", userA), userB); err != nil {
						nc.logger.Error("ZREM Error", zap.Error(err))
					}

					//在A的移除好友列表里删除B ZREM
					if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:2", userA), userB); err != nil {
						nc.logger.Error("ZREM Error", zap.Error(err))
					}

					//在B的移除好友列表里删除A ZREM
					if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:2", userB), userA); err != nil {
						nc.logger.Error("ZREM Error", zap.Error(err))
					}

					//直接让双方成为好友
					if _, err = redisConn.Do("ZADD", fmt.Sprintf("Friend:%s:1", userA), time.Now().UnixNano()/1e6, userB); err != nil {
						nc.logger.Error("ZADD Error", zap.Error(err))
					}
					if _, err = redisConn.Do("ZADD", fmt.Sprintf("Friend:%s:1", userB), time.Now().UnixNano()/1e6, userA); err != nil {
						nc.logger.Error("ZADD Error", zap.Error(err))
					}

					//如果userB是商户
					if userType == int(User.UserType_Ut_Business) {
						//在用户的关注有序列表里增加此商户
						if _, err = redisConn.Do("ZADD", fmt.Sprintf("Watching:%s", userA), time.Now().UnixNano()/1e6, userB); err != nil {
							nc.logger.Error("ZADD Error", zap.Error(err))
						}
						//更新redis的sync:{用户账号} watchAt 时间戳
						redisConn.Do("HSET",
							fmt.Sprintf("sync:%s", userA),
							"watchAt",
							time.Now().UnixNano()/1e6)
					}

					//增加A的好友B的信息哈希表
					nick, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", userB), "Nick"))
					_, err = redisConn.Do("HMSET",
						fmt.Sprintf("FriendInfo:%s:%s", userA, userB),
						"Username", userB,
						"Nick", nick,
						"Source", req.GetSource(), //来源
						"Ex", req.GetPs(), //附言

					)

					//增加B的好友A的信息哈希表
					//HMSET FriendInfo:{B}:{A} username {username} nick {nick} source {source} ex {ex} createAt {createAt} updateAt {updateAt}
					nick, _ = redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", userA), "Nick"))
					_, err = redisConn.Do("HMSET",
						fmt.Sprintf("FriendInfo:%s:%s", userB, userA),
						"Username", userA,
						"Nick", nick,
						"Source", req.GetSource(), //来源
						"Ex", req.GetPs(), //附言
					)

					//写入MySQL的好友表, 需要增加两条记录
					{

						userID_A, _ := redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("userData:%s", userA), "ID"))
						userID_B, _ := redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("userData:%s", userB), "ID"))

						pFriendA := new(models.Friend)
						pFriendA.UserID = userID_A
						pFriendA.FriendUserID = userID_B
						pFriendA.FriendUsername = userB
						if err := nc.service.AddFriend(pFriendA); err != nil {
							nc.logger.Error("Add Friend Error", zap.Error(err))
							errorCode = LMCError.AddFriendError
							goto COMPLETE
						}

						pFriendB := new(models.Friend)
						pFriendB.UserID = userID_B
						pFriendB.FriendUserID = userID_A
						pFriendB.FriendUsername = userA
						if err := nc.service.AddFriend(pFriendB); err != nil {
							nc.logger.Error("Add Friend Error", zap.Error(err))
							errorCode = LMCError.DataBaseError
							goto COMPLETE
						}

					}

					//下发MessageNotification通知给A所有端
					{
						//构造回包里的数据
						if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", userA))); err != nil {
							nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
							errorCode = LMCError.RedisError

							goto COMPLETE
						}
						body := &Msg.MessageNotificationBody{
							Type:           Msg.MessageNotificationType_MNT_PassFriendApply, //userB同意加你为好友
							HandledAccount: userB,
							HandledMsg:     "同意加你为好友",
							Status:         Msg.MessageStatus_MOS_Passed, //消息状态: 已通过验证
							Data:           psSourceData,                 // 用来存储附言及来源
							To:             userA,
						}
						bodyData, _ := proto.Marshal(body)
						eRsp := &Msg.RecvMsgEventRsp{
							Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
							Type:         Msg.MessageType_MsgType_Notification, //通知类型
							Body:         bodyData,
							From:         userB,                              //谁发的
							FromDeviceId: deviceID,                           //哪个设备发的
							Recv:         userA,                              //接收方, 根据场景判断Recv是个人还是群
							ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
							WorkflowID:   "",                                 //工作流ID
							Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
							Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
							Time:         uint64(time.Now().UnixNano() / 1e6),
						}
						go func() {
							time.Sleep(100 * time.Millisecond)
							nc.logger.Debug("5-2, 对方设置了运行任何人加好友",
								zap.String("userB", userB),
								zap.String("to", userA),
							)
							nc.BroadcastSystemMsgToAllDevices(eRsp, userA)
						}()

					}

					//下发通知给B所有端
					// 例外： A好友列表中没有B，B好友列表有A，A发起好友申请，A会收到B好友通过系统通知，B不接收好友申请系统通知。
					if !isBhaveA {
						//构造回包里的数据
						if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", userB))); err != nil {
							nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
							errorCode = LMCError.RedisError
							goto COMPLETE
						}

						body := &Msg.MessageNotificationBody{
							Type:           Msg.MessageNotificationType_MNT_PassFriendApply, //对方同意加你为好友
							HandledAccount: userA,
							HandledMsg:     "由于您设置了运行任何人加好友， 因此加好友成功",
							Status:         1,            // 消息状态  存储
							Data:           psSourceData, // 用来存储附言及来源
							To:             userB,
						}
						bodyData, _ := proto.Marshal(body)
						eRsp := &Msg.RecvMsgEventRsp{
							Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
							Type:         Msg.MessageType_MsgType_Notification, //通知类型
							Body:         bodyData,
							From:         userA,                              //谁发的
							FromDeviceId: deviceID,                           //哪个设备发的
							ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
							Recv:         userB,                              //接收方, 根据场景判断Recv是个人还是群
							WorkflowID:   "",                                 //工作流ID
							Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
							Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
							Time:         uint64(time.Now().UnixNano() / 1e6),
						}
						go func() {
							time.Sleep(100 * time.Millisecond)
							nc.logger.Debug("5-2,由于您设置了运行任何人加好友， 因此加好友成功",
								zap.String("to", userB),
							)
							nc.BroadcastSystemMsgToAllDevices(eRsp, userB)
						}()
					}

					//更新redis的sync:{用户账号} friendsAt 时间戳
					redisConn.Do("HSET",
						fmt.Sprintf("sync:%s", userA),
						"friendsAt",
						time.Now().UnixNano()/1e6)

					redisConn.Do("HSET",
						fmt.Sprintf("sync:%s", userB),
						"friendsAt",
						time.Now().UnixNano()/1e6)

				} else if allowType == LMCommon.NeedConfirm { //加好友设定是需要审核
					nc.logger.Debug("加好友设定是需要审核")

					//判断是否已经在预审核好友列表里
					if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Friend:%s:0", req.GetUsername()), username); err == nil {
						if reply != nil {
							nc.logger.Info("用户已经在预审核好友列表里", zap.String("Username", req.GetUsername()))
							//删除旧的
							_, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:0", req.GetUsername()), username)
						}
					}

					//redis里增加A的预审核好友列表, 注意: 好友请求有效期
					if _, err = redisConn.Do("ZADD", fmt.Sprintf("Friend:%s:0", userA), time.Now().UnixNano()/1e6, userB); err != nil {
						nc.logger.Error("ZADD Error", zap.Error(err))
					}

					//生成工作流ID
					workflowID := uuid.NewV4().String()
					rsp.Username = username
					rsp.Status = Friends.OpStatusType_Ost_WaitConfirm
					rsp.WorkflowID = workflowID

					//构造回包里的数据
					if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", userB))); err != nil {
						nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))

					}
					//HandledAccount 最后处理人
					//添加好友，对方接收/拒绝后，该字段填充为对方ID
					//申请入群，管理员通过/拒绝后，该字段填充管理员ID
					//邀请入群，用户通过/拒绝后，该字段填充目标用户ID
					body := &Msg.MessageNotificationBody{
						Type:           Msg.MessageNotificationType_MNT_ApplyFriend, //好友请求
						HandledAccount: userA,
						HandledMsg:     "添加好友请求",
						Status:         Msg.MessageStatus_MOS_Init, //未处理状态
						Data:           psSourceData,               // 用来存储附言及来源
						To:             userB,
					}
					bodyData, _ := proto.Marshal(body)
					eRsp := &Msg.RecvMsgEventRsp{
						Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
						Type:         Msg.MessageType_MsgType_Notification, //通知类型
						Body:         bodyData,
						From:         username,                           //谁发的
						FromDeviceId: deviceID,                           //哪个设备发的
						Recv:         userB,                              //接收方, 根据场景判断to是个人还是群
						ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
						WorkflowID:   workflowID,                         //服务端生成的工作流ID
						Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
						Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
						Time:         uint64(time.Now().UnixNano() / 1e6),
					}

					//A和B互相不为好友，B所有终端均会收到该消息。
					if !isAhaveB && !isBhaveA {
						//Go程，下发系统通知给B
						nc.logger.Debug("A和B互相不为好友，B所有终端均会收到该消息。下发系统通知给B", zap.String("userB", userB))

						go func() {
							time.Sleep(100 * time.Millisecond)
							nc.logger.Debug("5-2, 下发系统通知给userB",
								zap.String("to", userB),
							)
							nc.BroadcastSystemMsgToAllDevices(eRsp, userB)
						}()

					} else if isAhaveB && !isBhaveA { //A好友列表中有B，B好友列表没有A，A发起好友申请，B所有终端均会接收该消息，并且B可以选择同意、拒绝
						//Go程，下发系统通知给B
						nc.logger.Debug("A好友列表中有B，B好友列表没有A, 下发系统通知给B", zap.String("userB", userB))
						go func() {
							time.Sleep(100 * time.Millisecond)
							nc.logger.Debug("5-2, 下发系统通知给userB",
								zap.String("to", userB),
							)
							nc.BroadcastSystemMsgToAllDevices(eRsp, userB)
						}()

					} else if !isAhaveB && isBhaveA { //A好友列表中没有B，B好友列表有A，A发起好友申请，A会收到B好友通过系统通知，B不接收好友申请系统通知。
						nc.logger.Debug("A好友列表中没有B，B好友列表有A，A发起好友申请，A会收到B好友通过系统通知，B不接收好友申请系统通知。", zap.String("userA", userA))

						//TODO 直接向userA发好友已经通过
						{
							body := &Msg.MessageNotificationBody{
								Type:           Msg.MessageNotificationType_MNT_PassFriendApply, //userB同意加你为好友
								HandledAccount: userB,
								HandledMsg:     "B好友列表有A， 因此加好友成功",
								Status:         1,            // 消息状态  存储
								Data:           psSourceData, // 用来存储附言及来源
								To:             userA,
							}
							bodyData, _ := proto.Marshal(body)
							eRsp := &Msg.RecvMsgEventRsp{
								Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
								Type:         Msg.MessageType_MsgType_Notification, //通知类型
								Body:         bodyData,
								From:         userB,                              //谁发的
								FromDeviceId: deviceID,                           //哪个设备发的
								ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
								Recv:         userA,                              //接收方, 根据场景判断Recv是个人还是群
								WorkflowID:   "",                                 //工作流ID
								Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
								Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
								Time:         uint64(time.Now().UnixNano() / 1e6),
							}
							go func() {
								time.Sleep(100 * time.Millisecond)
								nc.logger.Debug("5-2, B好友列表有A， 因此加好友成功 下发系统通知给userA",
									zap.String("to", userA),
								)
								nc.BroadcastSystemMsgToAllDevices(eRsp, userA)
							}()

						}

					}

				}
			}
		case Friends.OptType_Fr_PassFriendApply: //userB同意加userA为好友
			{
				userA := req.GetUsername() //最开始的发起方
				userB := username          //当前用户

				//注意: 好友请求有效期 7 天， 定时任务模块要处理
				//判断是否已经在预审核好友列表里
				if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Friend:%s:0", userA), userB); err == nil {
					if reply == nil {
						nc.logger.Info("好友请求有效期 7 天, 用户已经不在预审核好友列表里", zap.String("Username", userB))
						errorCode = LMCError.AddFriendExpireError
						goto COMPLETE
					}
				}

				nc.logger.Debug(fmt.Sprintf("对方同意加你为好友, OptType_Fr_PassFriendApply, userA: %s, userB: %s", userA, userB))

				rsp.Username = username
				rsp.Status = Friends.OpStatusType_Ost_ApplySucceed

				//在A的预审核好友列表里删除B ZREM
				if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:0", userA), userB); err != nil {
					nc.logger.Error("ZREM Error", zap.Error(err))
				}

				//在A的移除好友列表里删除B ZREM
				if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:2", userA), userB); err != nil {
					nc.logger.Error("ZREM Error", zap.Error(err))
				}

				//在B的移除好友列表里删除A ZREM
				if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:2", userB), userA); err != nil {
					nc.logger.Error("ZREM Error", zap.Error(err))
				}

				//让双方成为好友
				if _, err = redisConn.Do("ZADD", fmt.Sprintf("Friend:%s:1", userA), time.Now().UnixNano()/1e6, userB); err != nil {
					nc.logger.Error("ZADD Error", zap.Error(err))
				}
				if _, err = redisConn.Do("ZADD", fmt.Sprintf("Friend:%s:1", userB), time.Now().UnixNano()/1e6, userA); err != nil {
					nc.logger.Error("ZADD Error", zap.Error(err))
				}

				//增加A的好友B的信息哈希表
				//HMSET FriendInfo:{A}:{B} username {username} nick {nick} source {source} ex {ex} createAt {createAt} updateAt {updateAt}
				nick, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", userB), "Nick"))
				_, err = redisConn.Do("HMSET",
					fmt.Sprintf("FriendInfo:%s:%s", userA, userB),
					"Username", userB,
					"Nick", nick,
					"Source", req.GetSource(),
					"Ex", "", //TODO
				)

				//增加B的好友A的信息哈希表
				//HMSET FriendInfo:{B}:{A} username {username} nick {nick} source {source} ex {ex} createAt {createAt} updateAt {updateAt}
				nick, _ = redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", userA), "Nick"))
				_, err = redisConn.Do("HMSET",
					fmt.Sprintf("FriendInfo:%s:%s", userB, userA),
					"Username", userA,
					"Nick", nick,
					"Source", req.GetSource(),
					"Ex", "", //TODO
				)

				//如果userB是商户
				if userType == int(User.UserType_Ut_Business) {
					//在用户的关注有序列表里增加此商户
					if _, err = redisConn.Do("ZADD", fmt.Sprintf("Watching:%s", userA), time.Now().UnixNano()/1e6, userB); err != nil {
						nc.logger.Error("ZADD Error", zap.Error(err))
					}
					//更新redis的sync:{用户账号} watchAt 时间戳
					redisConn.Do("HSET",
						fmt.Sprintf("sync:%s", userA),
						"watchAt",
						time.Now().UnixNano()/1e6)
				}

				//写入数据库，增加两条记录
				{

					userIDA, _ := redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("userData:%s", userA), "ID"))
					userIDB, _ := redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("userData:%s", userB), "ID"))

					pFriendA := new(models.Friend)
					pFriendA.UserID = userIDA
					pFriendA.FriendUserID = userIDB
					pFriendA.FriendUsername = userB
					if err := nc.service.AddFriend(pFriendA); err != nil {
						nc.logger.Error("Add Friend Error", zap.Error(err))
						errorCode = LMCError.AddFriendError
						goto COMPLETE
					}

					pFriendB := new(models.Friend)
					pFriendB.UserID = userIDB
					pFriendB.FriendUserID = userIDA
					pFriendB.FriendUsername = userA
					if err := nc.service.AddFriend(pFriendB); err != nil {
						nc.logger.Error("Add Friend Error", zap.Error(err))
						errorCode = LMCError.AddFriendError
						goto COMPLETE
					}

				}

				//更新redis的sync:{用户账号} friendsAt 时间戳
				redisConn.Do("HSET",
					fmt.Sprintf("sync:%s", userA),
					"friendsAt",
					time.Now().UnixNano()/1e6)

				redisConn.Do("HSET",
					fmt.Sprintf("sync:%s", userB),
					"friendsAt",
					time.Now().UnixNano()/1e6)

				//下发通知给A所有端
				{
					//构造回包里的数据
					if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", userA))); err != nil {
						nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					}

					body := &Msg.MessageNotificationBody{
						Type:           Msg.MessageNotificationType_MNT_PassFriendApply, //对方同意加你为好友
						HandledAccount: userB,
						HandledMsg:     "对方同意加你为好友",
						Status:         1,
						Data:           psSourceData, // 用来存储附言及来源
						To:             userA,
					}
					bodyData, _ := proto.Marshal(body)
					eRsp := &Msg.RecvMsgEventRsp{
						Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
						Type:         Msg.MessageType_MsgType_Notification, //通知类型
						Body:         bodyData,
						From:         username,                           //谁发的
						FromDeviceId: deviceID,                           //哪个设备发的
						ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
						Recv:         userA,                              //接收方, 根据场景判断to是个人还是群
						WorkflowID:   "",                                 //工作流ID
						Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
						Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
						Time:         uint64(time.Now().UnixNano() / 1e6),
					}
					isSend := false

					if !isAhaveB && !isBhaveA { //A和B互相不是好友，B通过/拒绝申请后,A所有终端会收到该系统通知。
						isSend = true
					} else if isAhaveB && !isBhaveA { //A好友列表中有B，B好友列表没有A，A发起好友申请，B通过或拒绝后，A接收该系统通知。
						isSend = true
					} else if !isAhaveB && isBhaveA { //A好友列表中没有B，B好友列表有A，A发起好友申请，服务端无须发送系统消息给B，而是直接给A接收好友通过事件
						isSend = true
					}
					if isSend {
						nc.logger.Debug(fmt.Sprintf("好友添加成功，下发通知给A, userA: %s, userB: %s", userA, userB))

						go func() {
							time.Sleep(100 * time.Millisecond)
							nc.logger.Debug("5-2, 对方同意加你为好友",
								zap.String("userB", userB),
								zap.String("to", userA),
							)
							nc.BroadcastSystemMsgToAllDevices(eRsp, userA)
						}()

					} else {
						nc.logger.Warn(fmt.Sprintf("警告: isSend=false, userA: %s, userB: %s", userA, userB))
					}
				}

			}

		case Friends.OptType_Fr_RejectFriendApply: //对方拒绝添加好友
			{
				userA := req.GetUsername() //发起方
				userB := username          //此时username是被加的人

				//注意: 好友请求有效期 7 天， 定时任务模块要处理
				//判断是否已经在预审核好友列表里
				if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Friend:%s:0", userA), userB); err == nil {
					if reply == nil {
						nc.logger.Info("好友请求有效期 7 天, 用户已经不在预审核好友列表里", zap.String("Username", userB))
						errorCode = LMCError.AddFriendExpireError
						goto COMPLETE
					}
				}

				nc.logger.Debug(fmt.Sprintf("对方拒绝添加好友, OptType_Fr_RejectFriendApply, userA: %s, userB: %s", userA, userB))

				rsp.Username = username
				rsp.Status = Friends.OpStatusType_Ost_RejectFriendApply

				//在A的预审核好友列表里删除B ZREM
				if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:0", userA), userB); err != nil {
					nc.logger.Error("ZREM Error", zap.Error(err))
				}

				//下发通知给A所有端
				{
					//构造回包里的数据
					if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", userA))); err != nil {
						nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					}

					body := &Msg.MessageNotificationBody{
						Type:           Msg.MessageNotificationType_MNT_RejectFriendApply, //对方拒绝添加好友
						HandledAccount: userB,
						HandledMsg:     "对方拒绝添加好友",
						Status:         1,            //TODO, 消息状态  存储
						Data:           psSourceData, // 用来存储附言及来源
						To:             userA,
					}
					bodyData, _ := proto.Marshal(body)
					eRsp := &Msg.RecvMsgEventRsp{
						Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
						Type:         Msg.MessageType_MsgType_Notification, //通知类型
						Body:         bodyData,
						From:         username,                           //谁发的
						FromDeviceId: deviceID,                           //哪个设备发的
						ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
						Recv:         userA,                              //接收方, 根据场景判断to是个人还是群
						WorkflowID:   "",                                 //工作流ID
						Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
						Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
						Time:         uint64(time.Now().UnixNano() / 1e6),
					}
					nc.logger.Debug(fmt.Sprintf("下发通知给A, userA: %s, userB: %s", userA, userB))

					go func() {
						time.Sleep(100 * time.Millisecond)
						nc.logger.Debug("5-2, 向userA推送拒绝添加好友的回复",
							zap.String("userB", userB),
							zap.String("to", userA),
						)
						nc.BroadcastSystemMsgToAllDevices(eRsp, userA)
					}()

				}

			}
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ = proto.Marshal(rsp)
		msg.FillBody(data) //网络包的body，承载真正的业务数据
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
3-5 移除好友,
A和B互为好友，A发起双向删除，则B所有在线终端会收到好友删除消息，from:A，to:B,默认只支持双向删除
*/
func (nc *NsqClient) HandleDeleteFriend(msg *models.Message) error {
	var err error
	errorCode := 200

	var newSeq uint64

	var isAhaveB, isBhaveA bool //A好友列表里有B， B好友列表里有A

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleDeleteFriend start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("FriendRequest",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Friends.DeleteFriendReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		targetUsername := req.GetUsername() //对方的用户账号
		nc.logger.Debug("FriendRequest body",
			zap.String("Username", targetUsername))

		//检测目标用户是否存在及添加好友的设定
		isExists, _ := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("userData:%s", targetUsername)))
		if !isExists {
			errorCode = LMCError.UserNotExistsError
			goto COMPLETE
		}

		//目标用户的用户类型
		userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", targetUsername), "UserType"))

		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Friend:%s:1", username), targetUsername); err == nil {
			if reply == nil {
				//A好友列表中没有B
				isAhaveB = false
			} else {
				isAhaveB = true
			}
		}

		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Friend:%s:1", targetUsername), username); err == nil {
			if reply == nil {
				//B好友列表中没有A
				isBhaveA = false
			} else {
				isBhaveA = true
			}

		}
		if isBhaveA {
			nc.logger.Debug(fmt.Sprintf("%s的好友列表有%s", targetUsername, username))
		}

		//本地好友表，删除双方的好友关系
		if !isAhaveB {
			err = nil
			errorCode = LMCError.IsNotFriendError
			goto COMPLETE
		}
		//在A的好友列表里删除B ZREM
		if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:1", username), targetUsername); err != nil {
			nc.logger.Error("ZREM Error", zap.Error(err))
		}

		//在B的好友列表里删除A ZREM
		if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:1", targetUsername), username); err != nil {
			nc.logger.Error("ZREM Error", zap.Error(err))
		}

		//如果B是商户
		if userType == int(User.UserType_Ut_Business) {
			//在用户的关注有序列表里删除 此商户
			if _, err = redisConn.Do("ZREM", fmt.Sprintf("Watching:%s", username), targetUsername); err != nil {
				nc.logger.Error("ZREM Error", zap.Error(err))
			}
			//增加用户的取消关注有序列表
			if _, err = redisConn.Do("ZADD", fmt.Sprintf("CancelWatching:%s", username), time.Now().UnixNano()/1e6, targetUsername); err != nil {
				nc.logger.Error("ZADD Error", zap.Error(err))
			}
			//更新redis的sync:{用户账号} watchAt 时间戳
			redisConn.Do("HSET",
				fmt.Sprintf("sync:%s", username),
				"watchAt",
				time.Now().UnixNano()/1e6)
		}

		//增加到各自的删除好友列表
		if _, err = redisConn.Do("ZADD", fmt.Sprintf("Friend:%s:2", username), time.Now().UnixNano()/1e6, targetUsername); err != nil {
			nc.logger.Error("ZADD Error", zap.Error(err))
		}
		if _, err = redisConn.Do("ZADD", fmt.Sprintf("Friend:%s:2", targetUsername), time.Now().UnixNano()/1e6, username); err != nil {
			nc.logger.Error("ZADD Error", zap.Error(err))
		}

		//删除数据库里的两条记录
		userID, _ := redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "ID"))
		targetUserID, _ := redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("userData:%s", targetUsername), "ID"))

		nc.service.DeleteFriend(userID, targetUserID)
		nc.service.DeleteFriend(targetUserID, userID)

		//更新redis的sync:{用户账号} friendsAt 时间戳
		redisConn.Do("HSET",
			fmt.Sprintf("sync:%s", username),
			"friendsAt",
			time.Now().UnixNano()/1e6)

		redisConn.Do("HSET",
			fmt.Sprintf("sync:%s", targetUsername),
			"friendsAt",
			time.Now().UnixNano()/1e6)

		//A和B互为好友，A发起双向删除，则B所有在线终端会收到好友删除的系统通知
		{
			//构造回包里的数据

			if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", targetUsername))); err != nil {
				nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
			}

			body := &Msg.MessageNotificationBody{
				Type:           Msg.MessageNotificationType_MNT_DeleteFriend, //删除好友
				HandledAccount: username,                                     //主动方，也就是用户自己
				HandledMsg:     "多端同步删除好友",
				Status:         Msg.MessageStatus_MOS_Done,
				Data:           []byte(""), // 附带的文本 该系统消息的文本
				To:             targetUsername,
			}
			bodyData, _ := proto.Marshal(body)
			eRsp := &Msg.RecvMsgEventRsp{
				Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
				Type:         Msg.MessageType_MsgType_Notification, //通知类型
				Body:         bodyData,
				From:         username,                           //谁发的
				FromDeviceId: deviceID,                           //哪个设备发的
				ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
				Recv:         targetUsername,                     //接收方, 根据场景判断to是个人还是群
				WorkflowID:   "",                                 //工作流ID
				Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
				Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
				Time:         uint64(time.Now().UnixNano() / 1e6),
			}

			go func() {
				time.Sleep(100 * time.Millisecond)
				nc.logger.Debug("5-2, 用户自己主动删除好友",
					zap.String("主动方", username),
					zap.String("to", targetUsername),
				)
				nc.BroadcastSystemMsgToAllDevices(eRsp, targetUsername)
			}()

		}

		// //向当前用户的其它端推送
		// {
		// 	if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", username))); err != nil {
		// 		nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
		// 		errorCode = LMCError.RedisError
		// 		goto COMPLETE
		// 	}

		// 	body := Msg.MessageNotificationBody{
		// 		Type:           Msg.MessageNotificationType_MNT_MultiDeleteFriend, //多端同步删除好友
		// 		HandledAccount: username,                                          //当前用户
		// 		HandledMsg:     "",
		// 		Status:         Msg.MessageStatus_MOS_Done,
		// 		Data:           []byte(""),
		// 		To:             targetUsername, //删除的好友
		// 	}
		// 	bodyData, _ := proto.Marshal(&body)
		// 	eRsp := &Msg.RecvMsgEventRsp{
		// 		Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
		// 		Type:         Msg.MessageType_MsgType_Notification, //通知类型
		// 		Body:         bodyData,
		// 		From:         username,
		// 		FromDeviceId: deviceID,
		// 		ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
		// 		Recv:         targetUsername,                     //接收方, 根据场景判断to是个人还是群
		// 		Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对teamMembere这个用户的通知序号
		// 		Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
		// 		Time:         uint64(time.Now().UnixNano() / 1e6),
		// 	}

		// 	go func() {
		// 		time.Sleep(100 * time.Millisecond)
		// 		nc.logger.Debug("延时100ms消息, 5-2",
		// 			zap.String("to", targetUsername),
		// 		)
		// 		nc.BroadcastSystemMsgToAllDevices(eRsp, username, deviceID)
		// 	}()

		// }
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//只需返回200即可
		msg.FillBody(nil)
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
3-6 刷新好友资料
*/
func (nc *NsqClient) HandleUpdateFriend(msg *models.Message) error {
	var err error
	errorCode := 200

	var data []byte
	rsp := &Friends.UpdateFriendRsp{}

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleUpdateFriend start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("UpdateFriend",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Friends.UpdateFriendReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		nc.logger.Debug(" body",
			zap.String("Username", req.GetUsername()))

		userID, _ := redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "ID"))

		targetUsername := req.GetUsername() //对方的用户账号

		//查出 targetUser 有效性，是否已经是好友，好友增加的设置等信息

		targetKey := fmt.Sprintf("userData:%s", targetUsername)

		isExists, _ := redis.Bool(redisConn.Do("EXISTS", targetKey))
		if !isExists {
			errorCode = LMCError.RedisError
			goto COMPLETE
		}

		//fields
		//查询出需要修改的好友信息
		pFriend := new(models.Friend)
		where := models.Friend{UserID: userID, FriendUsername: targetUsername}
		if err = nc.db.Model(pFriend).Where(&where).First(pFriend).Error; err != nil {
			nc.logger.Error("Query friend Error", zap.Error(err))
			errorCode = LMCError.DataBaseError
			goto COMPLETE
		}

		if alias, ok := req.Fields[1]; ok {
			//修改呢称
			pFriend.Alias = alias
		}

		if ex, ok := req.Fields[2]; ok {
			//修改呢称
			pFriend.Extend = ex

		}

		if err := nc.service.UpdateFriend(pFriend); err != nil {
			nc.logger.Error("修改好友资料失败", zap.Error(err))
			errorCode = LMCError.DataBaseError
			goto COMPLETE
		}

		rsp.TimeTag = uint64(time.Now().UnixNano() / 1e6)
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ = proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
3-8 增量同步好友列表
*/
func (nc *NsqClient) HandleGetFriends(msg *models.Message) error {
	var err error
	errorCode := 200

	rsp := &Friends.GetFriendsRsp{}
	var data []byte

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleGetFriends start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("GetFriends",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Friends.GetFriendsReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		nc.logger.Debug("GetFriends body",
			zap.Uint64("timeTag", req.GetTimeTag()))

		rsp = &Friends.GetFriendsRsp{
			TimeTag:      uint64(time.Now().UnixNano() / 1e6),
			Friends:      make([]*Friends.Friend, 0),
			RemovedUsers: make([]string, 0),
		}

		//从redis的有序集合查询出score大于req.GetTimeTag()的成员
		friends, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:1", username), req.GetTimeTag(), "+inf"))
		for _, friendUsername := range friends {

			nick, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Nick"))
			source, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Source"))
			ex, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Ex"))
			avatar, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", friendUsername), "Avatar"))
			if err != nil {
				nc.logger.Error("HGET Avatar error", zap.Error(err))
			}
			if (avatar != "") && !strings.HasPrefix(avatar, "http") {

				avatar = LMCommon.OSSUploadPicPrefix + avatar + "?x-oss-process=image/resize,w_50/quality,q_50"
			}

			rsp.Friends = append(rsp.Friends, &Friends.Friend{
				Username: friendUsername,
				Avatar:   avatar,
				Nick:     nick,
				Source:   source,
				Ex:       ex,
			})
		}
		//从redis里读取username的删除的好友列表
		RemoveFriends, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:2", username), req.GetTimeTag(), "+inf"))
		for _, friendUsername := range RemoveFriends {
			rsp.RemovedUsers = append(rsp.RemovedUsers, friendUsername)
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {

		data, _ = proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
3-9 关注商户
*/
func (nc *NsqClient) HandleWatching(msg *models.Message) error {
	var err error
	errorCode := 200

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleWatching start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("Watching",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Friends.WatchingReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		nc.logger.Debug("Watching body",
			zap.String("BusinessUsername", req.BusinessUsername),
		)

		//判断是否是商户账号
		if userType, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", req.BusinessUsername), "UserType")); err != nil {
			nc.logger.Error("redisConn HGET Error", zap.Error(err))
			errorCode = LMCError.RedisError

			goto COMPLETE
		} else {
			if userType != int(User.UserType_Ut_Business) {
				nc.logger.Debug("User is not business type", zap.String("BusinessUsername", req.BusinessUsername))
				errorCode = LMCError.IsNotBusinessUserError
				goto COMPLETE
			}

		}
		//判断是否封号，是否存在
		if state, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", req.BusinessUsername), "State")); err != nil {
			nc.logger.Error("redisConn HGET Error", zap.Error(err))
			errorCode = LMCError.RedisError

			goto COMPLETE
		} else {
			if state == LMCommon.UserBlocked {
				nc.logger.Debug("User is blocked", zap.String("BusinessUsername", req.BusinessUsername))
				errorCode = LMCError.UserIsBlockedError
				goto COMPLETE
			}

		}

		//判断是否被对方拉黑
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("BlackList:%s:1", req.BusinessUsername), username); err == nil {
			if reply != nil {
				nc.logger.Warn("用户已被对方拉黑,不能关注", zap.String("Username", req.BusinessUsername))
				errorCode = LMCError.FollowIsBlackUserError
				goto COMPLETE
			}
		}

		//在用户的关注有序列表里增加此商户
		if _, err = redisConn.Do("ZADD", fmt.Sprintf("Watching:%s", username), time.Now().UnixNano()/1e6, req.BusinessUsername); err != nil {
			nc.logger.Error("ZADD Error", zap.Error(err))
		}
		//更新redis的sync:{用户账号} watchAt 时间戳
		redisConn.Do("HSET",
			fmt.Sprintf("sync:%s", username),
			"watchAt",
			time.Now().UnixNano()/1e6)

		//在商户的被关注有序列表里增加此用户
		if _, err = redisConn.Do("ZADD", fmt.Sprintf("BeWatching:%s", req.BusinessUsername), time.Now().UnixNano()/1e6, username); err != nil {
			nc.logger.Error("ZADD Error", zap.Error(err))
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		rsp := &Friends.WatchingRsp{
			BusinessUsername: req.BusinessUsername,
		}
		data, _ := proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
3-10 取消关注商户
*/
func (nc *NsqClient) HandleCancelWatching(msg *models.Message) error {
	var err error
	errorCode := 200

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleCancelWatching start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("CancelWatching",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Friends.CancelWatchingReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		nc.logger.Debug("CancelWatching body",
			zap.String("BusinessUsername", req.BusinessUsername),
		)

		//在用户的关注有序列表里移除此商户
		if _, err = redisConn.Do("ZREM", fmt.Sprintf("Watching:%s", username), req.BusinessUsername); err != nil {
			nc.logger.Error("ZREM Error", zap.Error(err))
		}

		//增加用户的取消关注有序列表
		if _, err = redisConn.Do("ZADD", fmt.Sprintf("CancelWatching:%s", username), time.Now().UnixNano()/1e6, req.BusinessUsername); err != nil {
			nc.logger.Error("ZADD Error", zap.Error(err))
		}
		//更新redis的sync:{用户账号} watchAt 时间戳
		redisConn.Do("HSET",
			fmt.Sprintf("sync:%s", username),
			"watchAt",
			time.Now().UnixNano()/1e6)

		//在商户的关注有序列表里移除此用户
		if _, err = redisConn.Do("ZREM", fmt.Sprintf("BeWatching:%s", req.BusinessUsername), username); err != nil {
			nc.logger.Error("ZREM Error", zap.Error(err))
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		rsp := &Friends.CancelWatchingRsp{
			BusinessUsername: req.BusinessUsername,
		}
		data, _ := proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}
