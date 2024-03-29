package models

import (
	"time"

	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"
)

//redis里订单数据
type OrderInfo struct {
	OrderID          string  `json:"order_id"`                              //订单ID
	TicketCode       uint64  `json:"ticket_code"`                           //出票码, redis里递增
	ProductID        string  `json:"product_id"`                            //商品ID
	Attach           string  `json:"attach"`                                //订单内容密文
	AttachHash       string  `json:"attach_hash"`                           //订单内容哈希
	BuyerUsername    string  `json:"buyer_username" validate:"required"`    //买家注册号
	OpkBuyUser       string  `json:"opk_buyeruser" validate:"required"`     //买家公钥
	BusinessUsername string  `json:"business_username" validate:"required"` //商户注册号
	OpkBusinessUser  string  `json:"opk_businessuser" validate:"required"`  //商户公钥
	Cost             float64 `json:"cost" validate:"required"`              //本订单的总金额
	State            int     `json:"state"`                                 //订单类型
	IsPayed          bool    `json:"is_payed"`                              //此订单支付状态， true- 支付完成，false-未支付
	IsUrge           bool    `json:"is_urge"`                               //催单
	BodyType         int     `json:"body_type"`                             //彩票类型
	BodyObjFile      string  `json:"body_objfile"`                          //订单body的rsa加密阿里云obj
	OrderImageFile   string  `json:"order_imagefile"`                       //订单拍照图片的阿里云obj
	BlockNumber      uint64  `json:"block_number"`                          //图片上链成功执行合约的所在区块高度
	TxHash           string  `json:"tx_hash" `                              //图片上链交易哈希
}

/*
服务端的订单图片上链历史表
*/
type OrderImagesHistory struct {
	global.LMC_Model

	OrderID          string  `gorm:"primarykey" json:"order_id"`            //订单ID
	ProductID        string  ` json:"product_id"`                           //商品ID
	BuyUsername      string  `json:"buy_username" validate:"required"`      //买家注册号
	BusinessUsername string  `json:"business_username" validate:"required"` //商户注册号
	Cost             float64 `json:"cost" validate:"required"`              //本订单的总金额

	//订单图片在商户的oss objectID
	BusinessOssImage string `json:"business_image" validate:"required"`

	//订单body加密阿里云obj
	BodyObjFile string ` json:"body_objfile"`

	BlockNumber uint64 `json:"block_number"` //成功执行合约的所在区块高度
	TxHash      string `json:"tx_hash" `     //交易哈希

}

//BeforeCreate CreatedAt赋值
func (o *OrderImagesHistory) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (o *OrderImagesHistory) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

type OrderItems struct {
	OrderId        string         `json:"order_id" gorm:"primarykey;type:char(36)"`
	CreatedAt      int64          `json:"created_at"`
	UpdatedAt      int64          `json:"updated_at" json:"-"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	StoreId        string         `json:"store_id"`
	UserId         string         `json:"user_id"`
	WitnessId      string         `json:"witness_id"`
	ProductId      string         `json:"product_id"`
	ChainAddress   string         `json:"chain_address"`
	Body           string         `json:"body"`
	PublicKey      string         `json:"public_key"`
	StorePublicKey string         `json:"store_public_key"`
	OrderStatus    int            `json:"order_status"`
	PayStatus      int            `json:"pay_status"`
	ChainStatus    int            `json:"chain_status"`
	Amounts        float64        `json:"amounts"`
	Fee            float64        `json:"fee"`
	CouponID       string         `json:"coupon_id"`
	ImageHash      string         `json:"image_hash"`
	TicketCode     int64          `json:"ticket_code"`  //出票码, redis里递增
	Prize          float64        `json:"prize"`        //中奖金额
	PrizedPhoto    string         `json:"prized_photo"` // 兑奖的图片
	//StoreInfo    User           `json:"-" gorm:"foreignKey:StoreId;references:Username" `
	//UserInfo     User           `json:"-" gorm:"foreignKey:UserId;references:Username" `
}

//BeforeCreate CreatedAt赋值
func (o *OrderItems) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (o *OrderItems) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//  用于记录订单支付的日志记录
type PayOrderLogItem struct {
	gorm.Model
	OperationType int     `json:"operation_type"` // 操作的类型 用户操作 , 系统回调 , 外部回调, 客服操作 等
	OrderID       string  `json:"order_id"`       // 处理的订单id
	OrderType     int     `json:"order_type"`     // 订单类型
	UserID        string  `json:"user_id"`        // 执行操作的用户id
	Log           string  `json:"log"`            // 日志信息
	Money         float64 `json:"money"`          // 操作的金额
	MoneyFrom     string  `json:"money_from"`     // 金额 来源流向
	MoneyTo       string  `json:"money_to"`       // 金额流向
}

/*
message UploadOrderBodyReq {
    //订单ID
    string orderID =1;
    //彩票类型 1-福彩 2-体彩
    int32 body_type =2;
    //订单附件里的body加密数据文件，已经上传到了aliyunoss
    string bodyObjFile =3;

}

message UploadOrderBodyResp  {
    //时间
    fixed64 time=4;
}
*/

//仅在商户端: 将订单body经过RSA加密后提交阿里云后，服务端会缓存在redis里
type UploadOrderBodyReq struct {

	//订单ID
	OrderID string `json:"orderID,omitempty"`

	//订单附件里的body加密数据文件，已经上传到了aliyunoss
	BodyObjFile string `json:"bodyObjFile,omitempty"`
}

//用于不定多条件查询订单
type ReqKeyWordDataType struct {
	TicketCode uint64 `json:"ticket_code"` //出票码, 精确匹配
	OrderID    string `json:"order_id"`    //订单ID, 模糊查询
	Limit      int    `json:"limit"`       //每页记录数
	Offset     int    `json:"offset"`      //分页
	StartTime  int64  `json:"start_time"`  //开始时间   不能为0
	EndTime    int64  `json:"end_time"`    //结束时间   不能为0
	Status     int    `json:"status"`      //订单状态
}

/*

{
	"ticket_code": 1, //出票码, 精确匹配
	"order_id": "a1",  //订单ID, 模糊查询
	"limit" : 20,  //每页记录数
	"offset" : 0,  //页数
	"start_time" : 162097313112,  //开始时间  如果非零，则结束时间也必须非零。如果是0，则忽略
	"end_time" : 163235533666,  //结束时间  如果非零，则开始时间也必须非零。如果是0，则忽略
	"status" : 17  //订单状态
}

*/
