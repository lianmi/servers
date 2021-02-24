/*
这个文件是和前端相关的restful接口-用户模块，/v1/store/....
*/
package controllers

import (
	"net/http"
	// "time"

	Order "github.com/lianmi/servers/api/proto/order"
	User "github.com/lianmi/servers/api/proto/user"
	"go.uber.org/zap"

	jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/common/codes"
	// "github.com/lianmi/servers/internal/pkg/models"
	// uuid "github.com/satori/go.uuid"
	// "go.uber.org/zap"
)

//根据商户注册id获取店铺资料
func (pc *LianmiApisController) GetStore(c *gin.Context) {
	code := codes.InvalidParams
	businessUsername := c.Param("id")

	if businessUsername == "" {
		RespData(c, http.StatusOK, 500, "id is empty")
		return
	}

	store, err := pc.service.GetStore(businessUsername)
	if err != nil {
		RespData(c, http.StatusOK, 400, err.Error())
		return
	}

	code = codes.SUCCESS
	RespData(c, http.StatusOK, code, store)
}

//TODO 返回商品种类 在新的数据库需要建立字典表
func (pc *LianmiApisController) GetStoreTypes(c *gin.Context) {

	type StoreTypeData struct {
		StoreType int    //编号
		Name      string //名称
	}
	storeTypes := make([]StoreTypeData, 0)

	storeTypes = append(storeTypes, StoreTypeData{
		StoreType: 1,
		Name:      "生鲜家禽",
	})

	storeTypes = append(storeTypes, StoreTypeData{
		StoreType: 2,
		Name:      "肉类",
	})

	storeTypes = append(storeTypes, StoreTypeData{
		StoreType: 3,
		Name:      "水果蔬菜",
	})

	storeTypes = append(storeTypes, StoreTypeData{
		StoreType: 4,
		Name:      "粮油食杂",
	})

	storeTypes = append(storeTypes, StoreTypeData{
		StoreType: 5,
		Name:      "熟食",
	})

	storeTypes = append(storeTypes, StoreTypeData{
		StoreType: 6,
		Name:      "面包糕点",
	})

	storeTypes = append(storeTypes, StoreTypeData{
		StoreType: 7,
		Name:      "生活五金",
	})

	storeTypes = append(storeTypes, StoreTypeData{
		StoreType: 8,
		Name:      "家政保姆",
	})

	storeTypes = append(storeTypes, StoreTypeData{
		StoreType: 9,
		Name:      "彩票",
	})

	storeTypes = append(storeTypes, StoreTypeData{
		StoreType: 10,
		Name:      "搬运货运",
	})

	storeTypes = append(storeTypes, StoreTypeData{
		StoreType: 11,
		Name:      "电器维修",
	})

	storeTypes = append(storeTypes, StoreTypeData{
		StoreType: 12,
		Name:      "服务行业",
	})

	storeTypes = append(storeTypes, StoreTypeData{
		StoreType: 13,
		Name:      "其它",
	})

	code := codes.SUCCESS

	RespData(c, http.StatusOK, code, storeTypes)
}

//增加或修改店铺资料
func (pc *LianmiApisController) AddStore(c *gin.Context) {

	code := codes.InvalidParams

	claims := jwt_v2.ExtractClaims(c)
	username := claims[LMCommon.IdentityKey].(string)
	if username == "" {
		RespData(c, http.StatusOK, 500, "username is empty")
		return
	}

	var req User.Store
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {
		// if req.Province == "" || req.County == "" || req.City == "" || req.Street == "" || req.LegalPerson == "" || req.LegalIdentityCard == "" {
		// 	RespData(c, http.StatusOK, code, "商户地址信息必填")
		// 	return
		// }
		if req.BusinessUsername == "" {
			RespData(c, http.StatusOK, code, "商户注册账号id必填")
			return
		} else {
			if req.BusinessUsername != username {
				RespData(c, http.StatusOK, code, "商户注册账号非当前登录账号")
				return
			}
		}
		if req.Branchesname == "" {
			RespData(c, http.StatusOK, code, "商户店铺名称必填")
			return
		}
		if req.ContactMobile == "" {
			RespData(c, http.StatusOK, code, "联系手机必填")
			return
		}
		if req.ImageUrl == "" {
			RespData(c, http.StatusOK, code, "商户店铺外景图片必填")
			return
		}

		// if req.BusinessLicenseUrl == "" {
		// 	RespData(c, http.StatusOK, code, "营业执照url必填")
		// 	return
		// }
		// if req.Wechat == "" {
		// 	RespData(c, http.StatusOK, code, "微信必填")
		// 	return
		// }

		//保存或增加
		if err := pc.service.AddStore(&req); err != nil {
			pc.logger.Error("pc.service.AddStore error ", zap.Error(err))
			RespData(c, http.StatusNotAcceptable, code, err.Error())
		} else {
			code = codes.SUCCESS
			RespOk(c, http.StatusOK, code)
		}

	}

}

//修根据gps位置获取一定范围内的店铺列表
func (pc *LianmiApisController) QueryStoresNearby(c *gin.Context) {

	code := codes.InvalidParams
	var req Order.QueryStoresNearbyReq

	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {

		resp, err := pc.service.GetStores(&req)
		if err != nil {
			RespData(c, http.StatusOK, code, "获取店铺列表错误")
			return
		}

		RespData(c, http.StatusOK, 200, resp)
	}

}

//获取某个用户对所有店铺点赞情况, UI会保存在本地表里,  UI主动发起同步
func (pc *LianmiApisController) UserLikes(c *gin.Context) {
	code := codes.InvalidParams

	claims := jwt_v2.ExtractClaims(c)
	username := claims[LMCommon.IdentityKey].(string)
	if username == "" {
		RespData(c, http.StatusOK, 500, "username is empty")
		return
	}

	userLikes, err := pc.service.UserLikes(username)
	if err != nil {
		RespData(c, http.StatusOK, 500, err.Error())
		return
	}

	code = codes.SUCCESS
	RespData(c, http.StatusOK, code, userLikes)

}

//获取店铺的所有点赞用户列表
func (pc *LianmiApisController) StoreLikes(c *gin.Context) {
	code := codes.InvalidParams

	businessUsername := c.Param("id")

	if businessUsername == "" {
		RespData(c, http.StatusOK, 500, "id is empty")
		return
	}

	storeLikes, err := pc.service.StoreLikes(businessUsername)
	if err != nil {
		RespData(c, http.StatusOK, 500, err.Error())
		return
	}

	code = codes.SUCCESS
	RespData(c, http.StatusOK, code, storeLikes)

}

//对某个店铺点赞
func (pc *LianmiApisController) ClickLike(c *gin.Context) {
	code := codes.InvalidParams

	claims := jwt_v2.ExtractClaims(c)
	username := claims[LMCommon.IdentityKey].(string)
	if username == "" {
		RespData(c, http.StatusOK, 500, "username is empty")
		return
	}

	businessUsername := c.Param("id")

	if businessUsername == "" {
		RespData(c, http.StatusOK, 500, "id is empty")
		return
	}

	linkCount, err := pc.service.ClickLike(username, businessUsername)
	if err != nil {
		RespData(c, http.StatusOK, 500, err.Error())
		return
	}

	code = codes.SUCCESS
	RespData(c, http.StatusOK, code, linkCount)

}

//取消对某个店铺点赞
func (pc *LianmiApisController) DeleteClickLike(c *gin.Context) {
	code := codes.InvalidParams

	claims := jwt_v2.ExtractClaims(c)
	username := claims[LMCommon.IdentityKey].(string)
	if username == "" {
		RespData(c, http.StatusOK, 500, "username is empty")
		return
	}

	businessUsername := c.Param("id")

	if businessUsername == "" {
		RespData(c, http.StatusOK, 500, "id is empty")
		return
	}

	totalLikeCount, err := pc.service.DeleteClickLike(username, businessUsername)
	if err != nil {
		RespData(c, http.StatusOK, 500, err.Error())
		return
	}

	code = codes.SUCCESS
	RespData(c, http.StatusOK, code, totalLikeCount)

}

//获取各种彩票的开售及停售时刻
func (pc *LianmiApisController) QueryLotterySaleTimes(c *gin.Context) {
	code := codes.InvalidParams

	lotterySaleTimesRsp, err := pc.service.QueryLotterySaleTimes()
	if err != nil {
		RespData(c, http.StatusOK, 500, err.Error())
		return
	}

	code = codes.SUCCESS
	RespData(c, http.StatusOK, code, lotterySaleTimesRsp)

}

//清除所有OPK
func (pc *LianmiApisController) ClearAllOPK(c *gin.Context) {
	code := codes.InvalidParams

	claims := jwt_v2.ExtractClaims(c)
	username := claims[LMCommon.IdentityKey].(string)
	if username == "" {
		RespData(c, http.StatusOK, 500, "username is empty")
		return
	}

	err := pc.service.ClearAllOPK(username)
	if err != nil {
		RespData(c, http.StatusOK, 500, err.Error())
		return
	}

	code = codes.SUCCESS
	RespOk(c, http.StatusOK, code)

}

//获取当前商户的所有OPK
func (pc *LianmiApisController) GetAllOPKS(c *gin.Context) {
	code := codes.InvalidParams

	claims := jwt_v2.ExtractClaims(c)
	username := claims[LMCommon.IdentityKey].(string)
	if username == "" {
		RespData(c, http.StatusOK, 500, "username is empty")
		return
	}

	resp, err := pc.service.GetAllOPKS(username)
	if err != nil {
		RespData(c, http.StatusOK, 500, err.Error())
		return
	}

	code = codes.SUCCESS
	RespData(c, http.StatusOK, code, resp)

}

//删除当前商户的指定OPK
func (pc *LianmiApisController) EraseOPK(c *gin.Context) {
	code := codes.InvalidParams

	claims := jwt_v2.ExtractClaims(c)
	username := claims[LMCommon.IdentityKey].(string)
	if username == "" {
		RespData(c, http.StatusOK, 500, "username is empty")
		return
	}

	var req Order.EraseOPKSReq

	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {

		err := pc.service.EraseOPK(username, &req)
		if err != nil {
			RespData(c, http.StatusOK, code, "删除当前商户的指定OPK错误")
			return
		}
		code = codes.SUCCESS
		RespOk(c, http.StatusOK, code)
	}

}

//设置当前商户的默认OPK， 当OPK池为空，则需要用到此OPK
func (pc *LianmiApisController) DefaultOPK(c *gin.Context) {
	code := codes.InvalidParams

	claims := jwt_v2.ExtractClaims(c)
	username := claims[LMCommon.IdentityKey].(string)
	if username == "" {
		RespData(c, http.StatusOK, 500, "username is empty")
		return
	}

	var req Order.DefaultOPKReq

	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {
		if req.Opk == "" {
			pc.logger.Error("binding JSON error ")
			RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段:Opk")
		} else {
			pc.logger.Debug("DefaultOPK", zap.String("req.Opk", req.Opk))
		}
		err := pc.service.SetDefaultOPK(username, req.Opk)
		if err != nil {
			RespData(c, http.StatusOK, code, "设置当前商户的默认OPK错误")
			return
		}
		code = codes.SUCCESS
		RespOk(c, http.StatusOK, code)
	}

}
