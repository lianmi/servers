/*
定义消息结构，消息是由IM客户端通过mqtt发布到broker，然后由本程序消费，并拆包，以场景维度推送到内部kafka
后端的各种服务（单聊服务，群聊服务，加密私聊服务，应用服务，电商，赛事，股市，交易所，期货，游戏服务端）等
负责处理，并将结果发送回kafka，本程序消费这些消息，并根据目标topic及target，发布到mqtt broker，客户端
订阅了这些topic，就可以收到这些处理后的消息，从而完成闭环。
*/
package models

import (
	uuid "github.com/satori/go.uuid"
)

//消息头
type MessageHeader struct {
	// the message id
	ID string `json:"id,omitempty"`
	// message type
	Type string `json:"type,omitempty"`

	// the time of creating
	Timestamp int64 `json:"timestamp,omitempty"`
	//tag for other need
	Tag string `json:"tag,omitempty"`
}

//MessageRoute contains structure of message
type MessageRoute struct {
	// where the message come from, 这里用于存储业务模块名称，如："Auth", "Users" ...
	Source string `json:"source,omitempty"`
	// where the message will broadcasted to
	Group string `json:"group, omitempty"`
	// where the message come to，后端的具体业务模块Topic订阅
	Target string `json:"target,omitempty"`
	
	// deviceid 
	DeviceId string `json:"deviceid,omitempty"`	

	//SDK发到服务端的taskid
	TaskId uint32 `json:"taskid,omitempty"` 
	//SDK账号id
	// Account string `json:"account,omitempty"`  
	// what's the business type
	BusinessTypeName string `json:"businesstypename,omitempty"`
	BusinessType     uint32 `json:"businesstype,omitempty"`
	BusinessSubType  uint32 `json:"businesssubtype,omitempty"`
}

// Message struct
type Message struct {
	Header  MessageHeader `json:"header"`
	Router  MessageRoute  `json:"router,omitempty"`

	//错误码， 200表示正常
	Code             int32  `json:"code,omitempty"`
	// 错误描述
	ErrorMsg []byte `json:"errormsg,omitempty"` 
	Content []byte        `json:"content,omitempty"`
}

//GetID returns message ID
func (msg *Message) GetID() string {
	return msg.Header.ID
}

//BuildRouter sets route and resource operation in message
func (msg *Message) BuildRouter(source, group, target string) *Message {
	msg.SetRoute(source, group, target)
	return msg
}

//SetBusinessAndSession sets router fields  in message
// func (msg *Message) SetBusinessAndSession(
// 											appKey string,
// 											os string,
// 										    clientType uint32,
// 										    clientVersion uint32,
// 										    clientMode uint32,
// 										    deviceId string,
// 										    cmdId uint32,
// 											taskId uint32, 
// 											businesstypename string, 
// 											businessType uint16, 
// 											businessSubType uint16,
// 										 ) *Message {

// 	msg.Router.AppKey = appKey
// 	msg.Router.Os = os
// 	msg.Router.ClientType = clientType
// 	msg.Router.ClientVersion = clientVersion
// 	msg.Router.ClientMode = clientMode
// 	msg.Router.DeviceId = deviceId
// 	msg.Router.CmdId = cmdId
// 	msg.Router.TaskId = taskId
// 	msg.Router.BusinessTypeName = businesstypename
// 	msg.Router.BusinessType = businessType
// 	msg.Router.BusinessSubType = businessSubType
// 	return msg
// }

func (msg *Message) SetDeviceId(deviceId string) *Message {
	msg.Router.DeviceId = deviceId
	return msg
}

func (msg *Message) GetDeviceId() string {
	return msg.Router.DeviceId
}

// func (msg *Message) SetCmdid(cmdId uint32) *Message {
// 	msg.Router.CmdId = cmdId
// 	return msg
// }

// func (msg *Message) GetCmdId() uint32 {
// 	return msg.Router.CmdId
// }

func (msg *Message) SetTaskId(taskId uint32) *Message {
	msg.Router.TaskId = taskId
	return msg
}

func (msg *Message) GetTaskId() uint32 {
	return msg.Router.TaskId
}

func (msg *Message) SetBusinessType(businesstype uint32) *Message {
	msg.Router.BusinessType = businesstype
	return msg
}

func (msg *Message) GetBusinessType() uint32 {
	return msg.Router.BusinessType
}

func (msg *Message) SetBusinessSubType(businesssubtype uint32) *Message {
	msg.Router.BusinessSubType = businesssubtype
	return msg
}

func (msg *Message) GetBusinessSubType() uint32 {
	return msg.Router.BusinessSubType
}

func (msg *Message) SetCode(code int32) *Message {
	msg.Code = code
	return msg
}

func (msg *Message) SetErrorMsg(errorMsg []byte) *Message {
	msg.ErrorMsg = errorMsg
	return msg
}

//SetRoute sets router source and group in message
func (msg *Message) SetRoute(source, group, target string) *Message {
	msg.Router.Source = source
	msg.Router.Group = group
	msg.Router.Target = target
	return msg
}

//SetTag set message tags
func (msg *Message) SetTag(tag string) *Message {
	msg.Header.Tag = tag
	return msg
}

//GetTag get message tags
func (msg *Message) GetTag() string {
	return msg.Header.Tag
}

//GetTimestamp returns message timestamp
func (msg *Message) GetTimestamp() int64 {
	return msg.Header.Timestamp
}

//GetContent returns message content
func (msg *Message) GetContent() []byte {
	return msg.Content
}

//GetBusinessTypeName returns message route BusinessTypeName
func (msg *Message) SetBusinessTypeName(businessTypeName string)  {
 	msg.Router.BusinessTypeName = businessTypeName
}

//GetBusinessTypeName returns message route BusinessTypeName
func (msg *Message) GetBusinessTypeName() string {
	return msg.Router.BusinessTypeName
}

func (msg *Message) GetCode() int32 {
	return msg.Code
}

func (msg *Message) GetErrorMsg() []byte {
	return msg.ErrorMsg
}

//GetSource returns message route source string
func (msg *Message) GetSource() string {
	return msg.Router.Source
}

//GetGroup returns message route group
func (msg *Message) GetGroup() string {
	return msg.Router.Group
}

//GetTarget returns message route Target
func (msg *Message) GetTarget() string {
	return msg.Router.Target
}

//UpdateID returns message object updating its ID
func (msg *Message) UpdateID() *Message {
	msg.Header.ID = uuid.NewV4().String()
	return msg
}

// BuildHeader builds message header. You can also use for updating message header
func (msg *Message) BuildHeader(typ string, timestamp int64) *Message {
	msg.Header.Type = typ
	msg.Header.Timestamp = timestamp
	return msg
}

//FillBody fills message  content that you want to send
func (msg *Message) FillBody(content []byte) *Message {
	msg.Content = content
	return msg
}

// NewRawMessage returns a new raw message:
// model.NewRawMessage().BuildHeader().BuildRouter().FillBody()
func NewRawMessage() *Message {
	return &Message{}
}

// NewMessage returns a new basic message:
// model.NewMessage().BuildRouter().FillBody()
func NewMessage(parentID string) *Message {
	msg := &Message{}
	msg.Header.ID = uuid.NewV4().String()
	return msg
}
