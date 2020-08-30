package kafkaBackend

import (
	"fmt"
	"net/http"

	"github.com/gomodule/redigo/redis"
	Order "github.com/lianmi/servers/api/proto/order"

	"github.com/lianmi/servers/internal/pkg/models"
	"google.golang.org/protobuf/proto"

	"go.uber.org/zap"
)

/*
1. 根据timeAt增量返回商品信息，首次timeAt请初始化为0，服务器返回全量商品信息，后续采取增量方式更新
2. 如果soldoutProducts不为空，终端根据soldoutProducts移除商品缓存数据
3. 获取商品信息的流程： 发起获取商品信息请求 → 更新本地数据库 → 返回数据给UI

*/
func (kc *KafkaClient) HandleQueryProducts(msg *models.Message) error {
	var err error
	// var toUser, teamID string
	errorCode := 200
	var errorMsg string
	// rsp := &Order.QueryProductsRsp{}

	// var newSeq uint64
	// var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleQueryProducts start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("QueryProducts",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Order.QueryProductsReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		kc.logger.Debug("QueryProducts  payload",
			zap.String("ProductId", req.GetProductId()),
			zap.Uint64("TimeAt", req.GetTimeAt()),
		)

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		// data, _ = proto.Marshal(rsp)
		// msg.FillBody(data)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("SendMsgRsp message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send SendMsgRsp message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}
