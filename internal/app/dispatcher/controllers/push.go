package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/common/codes"
	"github.com/lianmi/servers/internal/pkg/pushapi/huaweipush"
	"github.com/lianmi/servers/internal/pkg/pushapi/xiaomipush"
	"go.uber.org/zap"
)

//POST 方法 小米推送
func (pc *LianmiApisController) XiaomiPushTest(c *gin.Context) {
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

//POST 方法 华为推送
func (pc *LianmiApisController) HuaweiPushTest(c *gin.Context) {
	code := codes.InvalidParams
	type PushInfo struct {
		Title           string `json:"title"`           //通知栏推送消息标题
		Content         string `json:"content"`         //通知栏推送消息内容
		PushServiceCode string `json:"pushServiceCode"` //目标推送平台 0:谷歌,1:华为,2:小米,3:oppo,4:vivo,5:魅族,6:苹果
		AppId           string `json:"appId"`           //目标推送平台应用ID
		AppSecret       string `json:"appSecret"`       //密钥
		RegId           string `json:"regId"`           //推送的指定设备注册registration_id
		Class           string `json:"class,omitempty"` // 应用入口Activity类全路径。 样例：com.example.hmstest.MainActivity
	}

	var pushInfo PushInfo
	if c.BindJSON(&pushInfo) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
		return
	}
	if pushInfo.Title == "" {
		RespData(c, http.StatusOK, code, "Title为空")
		return
	}
	if pushInfo.Content == "" {
		RespData(c, http.StatusOK, code, "Content为空")
		return
	}

	//默认
	if pushInfo.AppId == "" {
		pushInfo.AppId = "104392783"
	}
	if pushInfo.AppSecret == "" {
		pushInfo.AppSecret = "098cf02491a1b050d6fdf4247cd2dbce9ad05d049bc4eeaee1da6721acc6220c"
	}
	if pushInfo.RegId == "" {
		pushInfo.RegId = "IQAAAACy0izgAAC-raDYmxHMMOvnBoHmnBFqJPShZ1-MdwZWTKjcm31FcXeBo9BeMh3iCgf_WIaEvbCAhpOiYxgA-MV5cg8doE2LAjSIHaWaqboViA"
	}

	//TODO: 华为推送
	client := huaweipush.NewClient(pushInfo.AppId, pushInfo.AppSecret)

	sendReq := &huaweipush.SendReq{
		Message: &huaweipush.Message{
			Android: &huaweipush.AndroidConfig{
				FastAppTarget: 2,
				Notification: &huaweipush.AndroidNotification{
					Title: pushInfo.Title,
					Body:  pushInfo.Content,
					ClickAction: &huaweipush.ClickAction{
						Type: 3,
					},
					Sound: strconv.Itoa(1),
					Badge: &huaweipush.BadgeNotification{
						AddNum: 1,
						Class:  pushInfo.Class,
					},
				},
			},
			Tokens: []string{pushInfo.RegId},
		},
	}
	sendRes, err := client.Send(sendReq)
	// fmt.Println(sendRes, err)
	if err != nil {
		RespData(c, http.StatusOK, 400, "向华为手机推送失败")
	}

	pc.logger.Debug("向华为手机", zap.Any("sendRes", sendRes))

	RespOk(c, http.StatusOK, 200)

	return
}
