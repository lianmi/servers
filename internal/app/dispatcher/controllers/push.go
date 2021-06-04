package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/common/codes"
	"github.com/modood/pushapi/xiaomipush"
	"go.uber.org/zap"
)

//POST 方法
func (pc *LianmiApisController) PushTest(c *gin.Context) {
	code := codes.InvalidParams
	type PushInfo struct {
		Title           string `json:"title"`           //通知栏推送消息标题
		Content         string `json:"content"`         //通知栏推送消息内容
		ChannelId       string `json:"channelId"`       //通知栏推送额外信息
		ChannelName     string `json:"channelName"`     //通知栏推送额外信息
		PushServiceCode string `json:"pushServiceCode"` //目标推送平台 0:谷歌,1:华为,2:小米,3:oppo,4:vivo,5:魅族,6:苹果
		AppSecret       string `json:"appSecret"`       //密钥
		RegId           string `json:"regId"`           //推送的指定设备注册registration_id
		AppId           string `json:"appId"`           //目标推送平台应用ID
	}

	var pushInfo PushInfo
	if c.BindJSON(&pushInfo) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
		return
	}

	//TODO: 小米推送
	client := xiaomipush.NewClient(pushInfo.AppSecret)

	sendReq := &xiaomipush.SendReq{
		RegistrationId: pushInfo.RegId,
		Title:          pushInfo.Title,
		Description:    pushInfo.Content,
		NotifyType:     2,
		Extra: &xiaomipush.Extra{
			NotifyEffect: "1",
			ChannelId:    pushInfo.ChannelId,
			ChannelName:  pushInfo.ChannelName,
		},
	}
	sendRes, err := client.Send(sendReq)
	// fmt.Println(sendRes, err)
	if err != nil {
		RespData(c, http.StatusOK, 400, "向小米推送失败")
	}

	pc.logger.Debug("向小米推送", zap.Any("sendRes", sendRes))

	RespOk(c, http.StatusOK, 200)

	return
}
