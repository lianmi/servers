package auth

import (
	"encoding/json"
)

//双色球基础数据结构
type ShuangSeQiu struct {
	DantuoBall []*int //胆拖
	RedBall    []*int //红球区
	BlueBall   []*int //篮球区
}

//双色球订单, 支持单式\复式\胆拖
type ShuangSeQiuOrder struct {
	Straw   []*ShuangSeQiu //用户选号后的数据
	Count   int            //总注数
	Cost    float64        //花费, 每注2元, 乘以总注数
	OrderID string
}

func (m *ShuangSeQiuOrder) ToJson() (string, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
