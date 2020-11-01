/*
这个文件是和前端相关的restful接口-UI显示:
/shops/....
*/
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	Auth "github.com/lianmi/servers/api/proto/auth"
)

//根据用户当前经纬度查询附近商店信息
func (pc *LianmiApisController) QueryShopsNearby(c *gin.Context) {
	/*
		claims := jwt_v2.ExtractClaims(c)
		userName := claims[common.IdentityKey].(string)
		deviceID := claims["deviceID"].(string)
		token := jwt_v2.GetToken(c)

		pc.logger.Debug("QueryShopsNearby",
			zap.String("userName", userName),
			zap.String("deviceID", deviceID),
			zap.String("token", token))

	*/

	resp := &Auth.QueryShopsNearbyResp{
		TotalPage: 1,
	}
	resp.Shops = append(resp.Shops, &Auth.Shop{
		BusinessUsername: "吴记牛肉店",    //商户账号id
		Nick:             "吴老板",      //商户呢称
		Avatar:           "",         //TODO, 阿里云的头像上传功能
		BranchesName:     "吴记牛肉店",    //商店名称
		Introductory:     "主营: 新鲜牛肉", //商店简介
		Province:         "广东省",      //省份
		City:             "广州市",      //城市
		County:           "天河区",      //区
		Street:           "棠下",       //街道
		Address:          "长堤路1号",    //地址
		Longitude:        0.32323,    //商店位置的经度
		Latitude:         2323,       //商店位置的经度
	})

	RespData(c, http.StatusOK, 200, resp)
}
