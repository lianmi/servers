/*
此接口是用于发送到nsq
*/
package controllers

import (
	"encoding/json"
	"time"

	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//通过nsq通道相关UI SDK发送消息 username, deviceID 必须是接收方的订阅
func (pc *LianmiApisController) SendMessagetoNsq(username, deviceID string, data []byte, businessType, businessSubType int) error {
	var err error
	if (username == "") || (deviceID == "") || (len(data) == 0) {
		return errors.Wrap(err, "username or deviceID or data is empty")
	}
	//向客户端响应 SyncFriendUsersEvent 事件
	targetMsg := &models.Message{}

	targetMsg.UpdateID()
	//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
	targetMsg.BuildRouter("Auth", "", "Auth.Frontend")

	targetMsg.SetJwtToken("")
	targetMsg.SetUserName(username) //
	targetMsg.SetDeviceID(deviceID) //
	// kickMsg.SetTaskID(uint32(taskId))
	targetMsg.SetBusinessTypeName("Auth")
	targetMsg.SetBusinessType(uint32(businessType))
	targetMsg.SetBusinessSubType(uint32(businessSubType))

	targetMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6)

	targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

	targetMsg.SetCode(200) //成功的状态码

	//构建数据完成，向dispatcher发送
	topic := "Auth.Frontend"
	rawData, _ := json.Marshal(targetMsg)
	if err = pc.nsqClient.Producer.Public(topic, rawData); err == nil {
		pc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		pc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
	}

	return err

}
