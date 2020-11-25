package order

import (
	"encoding/json"
)

//双色球基础数据结构
type ShuangSeQiu struct {
	DantuoBall []int //胆拖红球
	RedBall    []int //红球区
	BlueBall   []int //篮球区
}

//双色球订单, 支持单式\复式\胆拖
type ShuangSeQiuOrder struct {
	BuyUser          string         //买家
	BusinessUsername string         //商户注册账号
	ProductID        string         //商品ID
	OrderID          string         //订单ID
	LotteryPicHash   string         //彩票拍照的照片哈希
	LotteryPicURL    string         //彩票拍照的照片原图url
	LotteryType      int            //投注类型，1-单式\2-复式\3-胆拖
	Straws           []*ShuangSeQiu //用户选号后的数据，如果是单选，每个成员表示1注
	Count            int            //总注数
	Cost             float64        //花费, 每注2元, 乘以总注数
	TxHash           string         //上链的哈希
	BlockNumber      uint64         //区块高度
	CreatedAt        int64          //创建订单的时刻，服务端为准
	TakedAt          int64          //接单的时刻，服务端为准
	DoneAt           int64          //完成订单的时刻，服务端为准
	RefusedAt        int64          //拒单的时刻，服务端为准
}

func (m *ShuangSeQiuOrder) ToJson() (string, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
