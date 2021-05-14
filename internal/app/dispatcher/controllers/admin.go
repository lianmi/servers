/*
这个文件是和后台相关的restful接口，/admin/....
*/
package controllers

import (
	"strings"

	// "fmt"
	"net/http"
	// "strconv"
	// "time"

	// Global "github.com/lianmi/servers/api/proto/global"
	Auth "github.com/lianmi/servers/api/proto/auth"
	Order "github.com/lianmi/servers/api/proto/order"
	// User "github.com/lianmi/servers/api/proto/user"

	jwt_v2 "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"

	// "github.com/lianmi/servers/internal/app/dispatcher/services"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/common/codes"
	"github.com/lianmi/servers/internal/pkg/models"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
)

//检测用户是否有使用后台接口的权限
func (pc *LianmiApisController) CheckIsAdmin(c *gin.Context) bool {
	claims := jwt_v2.ExtractClaims(c)
	userName := claims[common.IdentityKey].(string)
	// deviceID := claims["deviceID"].(string)
	// token := jwt_v2.GetToken(c)

	// 用户是以admin开头的账号是后台管理账号
	if strings.HasPrefix(userName, "admin") {
		return true
	}

	return false

}

func (pc *LianmiApisController) CheckIsUser(c *gin.Context) (username, deviceid string, isok bool) {
	claims := jwt_v2.ExtractClaims(c)
	username = claims[common.IdentityKey].(string)
	deviceid = claims["deviceID"].(string)
	if username == "" || deviceid == "" {
		isok = false
		return
	}
	isok = true
	return
}

//封号
func (pc *LianmiApisController) BlockUser(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}
	pc.logger.Debug("BlockUser start ...")
	username := c.Param("username")
	if username == "" {
		RespData(c, http.StatusOK, 404, "Param is empty")
		return
	}

	err := pc.service.BlockUser(username)
	if err != nil {
		pc.logger.Error("Block User by username error", zap.Error(err))
		RespData(c, http.StatusOK, 500, "Block User by username error")

		return
	}

	RespOk(c, http.StatusOK, 200)
}

//解封
func (pc *LianmiApisController) DisBlockUser(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

	pc.logger.Debug("DisBlockUser start ...")
	username := c.Param("username")
	if username == "" {
		RespData(c, http.StatusOK, 404, "Param is empty")
		return
	}

	err := pc.service.DisBlockUser(username)
	if err != nil {
		pc.logger.Error("DisBlockUser User by username error", zap.Error(err))
		RespData(c, http.StatusOK, 404, "username  not found")
		return
	}

	RespOk(c, http.StatusOK, 200)
}

//后台: 授权新创建的群组
func (pc *LianmiApisController) ApproveTeam(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

	claims := jwt_v2.ExtractClaims(c)
	userName := claims[common.IdentityKey].(string)
	deviceID := claims["deviceID"].(string)
	token := jwt_v2.GetToken(c)

	pc.logger.Debug("ApproveTeam",
		zap.String("userName", userName),
		zap.String("deviceID", deviceID),
		zap.String("token", token))

	//读取
	teamID := c.Param("teamid")
	if teamID == "" {
		RespData(c, http.StatusOK, 400, "Param is empty")
		return
	}
	if err := pc.service.ApproveTeam(teamID); err == nil {
		pc.logger.Debug("ApproveTeam  run ok")
		RespOk(c, http.StatusOK, 200)
	} else {
		pc.logger.Debug("ApproveTeam run FAILD")
		RespData(c, http.StatusOK, 400, "授权新创建的群组失败")

	}

}

//封禁群组
func (pc *LianmiApisController) BlockTeam(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

	//读取
	teamID := c.Param("teamid")
	if teamID == "" {
		RespData(c, http.StatusOK, 400, "Param is empty")
		return
	}
	if err := pc.service.BlockTeam(teamID); err == nil {
		pc.logger.Debug("BlockTeam  run ok")
		RespOk(c, http.StatusOK, 200)
	} else {
		pc.logger.Debug("BlockTeam  run FAILD")
		RespData(c, http.StatusOK, 400, "封禁群组失败")

	}

}

//解封群组
func (pc *LianmiApisController) DisBlockTeam(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

	//读取
	teamID := c.Param("teamid")
	if teamID == "" {

		RespData(c, http.StatusOK, 400, "Param is empty")
		return
	}
	if err := pc.service.DisBlockTeam(teamID); err == nil {
		pc.logger.Debug("DisBlockTeam  run ok")
		RespOk(c, http.StatusOK, 200)
	} else {
		pc.logger.Debug("DisBlockTeam  run FAILD")
		RespData(c, http.StatusOK, 400, "封禁群组失败")

	}

}

func (pc *LianmiApisController) GetGeneralProductPage(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}
	code := codes.InvalidParams
	var req Order.GetGeneralProductPageReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {
		resp, err := pc.service.GetGeneralProductPage(&req)
		if err != nil {
			RespData(c, http.StatusOK, code, "获取店铺商品列表错误")
			return
		}

		RespData(c, http.StatusOK, 200, resp)

	}
}

//增加通用商品
func (pc *LianmiApisController) AddGeneralProduct(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

	var og Order.GeneralProduct
	if c.BindJSON(&og) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {
		//增加
		var productPic1Large, productPic2Large, productPic3Large string
		if len(og.ProductPics) >= 1 {
			productPic1Large = og.ProductPics[0].Large
		}
		if len(og.ProductPics) >= 2 {
			productPic2Large = og.ProductPics[1].Large
		}
		if len(og.ProductPics) >= 3 {
			productPic3Large = og.ProductPics[2].Large
		}

		generalProductInfo := &models.GeneralProductInfo{
			ProductId:        uuid.NewV4().String(), //商品ID
			ProductName:      og.ProductName,        //商品名称
			ProductType:      int(og.ProductType),   //商品种类枚举
			ProductDesc:      og.ProductDesc,        //商品详细介绍
			ProductPic1Large: productPic1Large,      //商品图片1-大图
			ProductPic2Large: productPic2Large,      //商品图片2-大图
			ProductPic3Large: productPic3Large,      //商品图片3-大图
			ShortVideo:       og.ShortVideo,         //商品短视频
			// AllowCancel:      *og.AllowCancel,        //是否允许撤单， 默认是可以，彩票类的不可以

		}

		if len(og.DescPics) >= 1 {
			generalProductInfo.DescPic1 = og.DescPics[0]
		}
		if len(og.DescPics) >= 2 {
			generalProductInfo.DescPic2 = og.DescPics[1]
		}
		if len(og.DescPics) >= 3 {
			generalProductInfo.DescPic3 = og.DescPics[2]
		}
		if len(og.DescPics) >= 4 {
			generalProductInfo.DescPic4 = og.DescPics[3]
		}
		if len(og.DescPics) >= 5 {
			generalProductInfo.DescPic5 = og.DescPics[4]
		}
		if len(og.DescPics) >= 6 {
			generalProductInfo.DescPic6 = og.DescPics[5]
		}

		if err := pc.service.AddGeneralProduct(generalProductInfo); err == nil {
			pc.logger.Debug("AddGeneralProduct run ok")
			// NOTE wujehy 添加成功 , 需要同时清理一下 缓存
			delete(pc.cacheMap, "CacheGetGeneralProjectIDs")
			delete(pc.cacheMap, "CacheGetGeneralProjectLists")
			RespOk(c, http.StatusOK, 200)
		} else {
			pc.logger.Warn("AddGeneralProduct run FAILD")
			RespData(c, http.StatusOK, 400, "增加通用商品失败")
		}

	}
}

//修改通用商品
func (pc *LianmiApisController) UpdateGeneralProduct(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		RespData(c, http.StatusOK, 401, "你不是管理员, 无权访问这个接口")
		return
	}

	//var og Order.GeneralProduct
	var og models.GeneralProductInfo
	if c.BindJSON(&og) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
		return
	} else {
		//修改
		if og.ProductId == "" {
			pc.logger.Warn("ProductId is empty")
			RespData(c, http.StatusOK, 400, "修改通用商品失败, ProductId 不能为空")
			return
		}

		//if len(og.ProductPics) == 0 {
		//	pc.logger.Warn("ProductPics length is 0")
		//	RespData(c, http.StatusOK, 400, "修改通用商品失败, ProductPics length is 0")
		//	return
		//}
		//var productPic1Large, productPic2Large, productPic3Large string
		//if len(og.ProductPics) >= 1 {
		//	productPic1Large = og.ProductPics[0].Large
		//}
		//if len(og.ProductPics) >= 2 {
		//	productPic2Large = og.ProductPics[1].Large
		//}
		//if len(og.ProductPics) >= 3 {
		//	productPic3Large = og.ProductPics[2].Large
		//}

		//generalProductInfo := &models.GeneralProductInfo{
		//	ProductId:        uuid.NewV4().String(), //商品ID
		//	ProductName:      og.ProductName,        //商品名称
		//	ProductType:      int(og.ProductType),   //商品种类枚举
		//	ProductDesc:      og.ProductDesc,        //商品详细介绍
		//	ProductPic1Large: productPic1Large,      //商品图片1-大图
		//	ProductPic2Large: productPic2Large,      //商品图片2-大图
		//	ProductPic3Large: productPic3Large,      //商品图片3-大图
		//	ShortVideo:       og.ShortVideo,         //商品短视频
		//	// AllowCancel:      *og.AllowCancel,        //是否允许撤单， 默认是可以，彩票类的不可以
		//
		//}
		//
		//if len(og.DescPics) >= 1 {
		//	generalProductInfo.DescPic1 = og.DescPics[0]
		//}
		//if len(og.DescPics) >= 2 {
		//	generalProductInfo.DescPic2 = og.DescPics[1]
		//}
		//if len(og.DescPics) >= 3 {
		//	generalProductInfo.DescPic3 = og.DescPics[2]
		//}
		//if len(og.DescPics) >= 4 {
		//	generalProductInfo.DescPic4 = og.DescPics[3]
		//}
		//if len(og.DescPics) >= 5 {
		//	generalProductInfo.DescPic5 = og.DescPics[4]
		//}
		//if len(og.DescPics) >= 6 {
		//	generalProductInfo.DescPic6 = og.DescPics[5]
		//}

		if err := pc.service.UpdateGeneralProduct(&og); err == nil {
			pc.logger.Debug("AddGeneralProduct  run ok")
			RespOk(c, http.StatusOK, 200)

			delete(pc.cacheMap, "CacheGetGeneralProjectIDs")
			delete(pc.cacheMap, "CacheGetGeneralProjectLists")

			return
		} else {
			pc.logger.Warn("AddGeneralProduct  run FAILD")
			RespData(c, http.StatusOK, 400, "修改通用商品失败")

			return
		}

	}
}

//删除通用商品
func (pc *LianmiApisController) DeleteGeneralProduct(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

	productId := c.Param("productid")
	if productId == "" {
		RespData(c, http.StatusOK, 404, "productid is empty")
		return
	}
	if pc.service.DeleteGeneralProduct(productId) {

		RespOk(c, http.StatusOK, 200)
	} else {
		RespData(c, http.StatusOK, 400, "delete GeneralProduct failed")
	}

}

//增加在线客服id
func (pc *LianmiApisController) AddCustomerService(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

	var req Auth.AddCustomerServiceReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {

		if !(req.Type == 1 || req.Type == 2) {
			RespData(c, http.StatusOK, 400, "Type参数错误")
		}
		if req.Username == "" {
			RespData(c, http.StatusOK, 400, "Username参数错误")
		}
		if req.JobNumber == "" {
			RespData(c, http.StatusOK, 400, "JobNumber参数错误")
		}
		if req.Evaluation == "" {
			RespData(c, http.StatusOK, 400, "Evaluation参数错误")
		}
		if req.NickName == "" {
			RespData(c, http.StatusOK, 400, "NickName参数错误")
		}

		err := pc.service.AddCustomerService(&req)

		if err != nil {
			RespData(c, http.StatusOK, 400, "Add CustomerService failed")
		} else {
			RespOk(c, http.StatusOK, 200)
		}
	}

}

//删除在线客服id
func (pc *LianmiApisController) DeleteCustomerService(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

	var req Auth.DeleteCustomerServiceReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {

		if req.Username == "" {
			RespData(c, http.StatusOK, 400, "Username参数错误")
		}

		if pc.service.DeleteCustomerService(&req) == false {
			RespData(c, http.StatusOK, 400, "Delete CustomerServices failed")
		} else {
			RespOk(c, http.StatusOK, 200)
		}
	}

}

//修改在线客服资料
func (pc *LianmiApisController) UpdateCustomerService(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

	var req Auth.UpdateCustomerServiceReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {
		if !(req.Type == 1 || req.Type == 2) {
			RespData(c, http.StatusOK, 400, "Type参数错误")
		}
		if req.Username == "" {
			RespData(c, http.StatusOK, 400, "Username参数错误")
		}
		if req.JobNumber == "" {
			RespData(c, http.StatusOK, 400, "JobNumber参数错误")
		}
		if req.Evaluation == "" {
			RespData(c, http.StatusOK, 400, "Evaluation参数错误")
		}
		if req.NickName == "" {
			RespData(c, http.StatusOK, 400, "NickName参数错误")
		}
		//修改在线客服资料
		err := pc.service.UpdateCustomerService(&req)

		if err != nil {
			RespData(c, http.StatusOK, 400, "Update CustomerServices failed")
		} else {
			RespOk(c, http.StatusOK, 200)
		}
	}

}

//将店铺审核通过
func (pc *LianmiApisController) AuditStore(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}
	var req Auth.AuditStoreReq

	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, 400, "参数错误, 缺少必填字段")
	} else {
		if req.BusinessUsername == "" {
			RespData(c, http.StatusOK, 400, "BusinessUsername参数错误")
		}

		err := pc.service.AuditStore(&req)

		if err != nil {
			RespData(c, http.StatusOK, 400, "AuditStore failed")
		} else {
			RespOk(c, http.StatusOK, 200)
		}
	}
}

//上传网点excel文件并导入到数据库里
func (pc *LianmiApisController) LoadExcel(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

	f, err := excelize.OpenFile("/tmp/upload/1619884694.xlsx")
	if err != nil {
		pc.logger.Error("OpenFile error ", zap.Error(err))
		return
	}
	rows, err := f.GetRows("Sheet1")
	for _, row := range rows {
		if len(row) >= 16 {
			storeType := 1
			keyword := row[0]
			storeName := row[6]
			if strings.Contains(keyword, "福彩") || strings.Contains(storeName, "福利") {
				storeType = 1
			} else {
				storeType = 2
			}
			lotteryStore := &models.LotteryStore{
				Keyword:   keyword,   //关键字 体彩 福彩
				MapID:     row[1],    //高德地图的id
				Province:  row[2],    //省份, 如广东省
				City:      row[3],    //城市，如广州市
				Area:      row[4],    //区，如天河区
				Address:   row[5],    //地址
				StoreName: storeName, //店铺名称
				Latitude:  row[8],    //商户地址的纬度
				Longitude: row[9],    //商户地址的经度
				Phones:    row[13],   //联系手机或电话, 以半角逗号隔开
				Photos:    row[15],   //店铺外景照片, 以半角逗号隔开
				StoreType: storeType, //店铺类型, 1-福彩 2-体彩
				Status:    0,         //状态，0-预，1-已提交
			}
			err := pc.service.SaveExcelToDb(lotteryStore)

			if err != nil {
				// RespData(c, http.StatusOK, 400, "SaveExcelToDb failed")
				pc.logger.Error("SaveExcelToDb error ", zap.Error(err))
			} else {
				// RespOk(c, http.StatusOK, 200)
				pc.logger.Debug("SaveExcelToDb ok ", zap.String("keyword", keyword), zap.String("storeName", storeName))
			}
		}
	}

	RespOk(c, http.StatusOK, 200)
}

//查询并分页获取采集的网点
func (pc *LianmiApisController) GetLotteryStores(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}

	req := models.LotteryStoreReq{}
	if err := c.BindQuery(&req); err != nil {
		req.Offset = 0
		req.Limit = 20
	}
	pc.logger.Debug("GetLotteryStores",
		zap.String("Keyword", req.Keyword),
		zap.String("Province", req.Province),
		zap.String("City", req.City),
		zap.String("Area", req.Area),
		zap.String("Address", req.Address), //模糊查找
	)

	//
	list, err := pc.service.GetLotteryStores(&req)
	if err != nil {
		RespFail(c, http.StatusOK, codes.InvalidParams, "未找到网点信息")
		return
	}

	RespData(c, http.StatusOK, codes.SUCCESS, list)
}

//管理员增加网点
func (pc *LianmiApisController) AdminAddStore(c *gin.Context) {
	if !pc.CheckIsAdmin(c) {
		return
	}
	code := codes.InvalidParams

	var req models.Store
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespData(c, http.StatusOK, code, "参数错误, 缺少必填字段")
	} else {
		// if req.Province == "" || req.Area == "" || req.City == "" || req.Street == "" || req.LegalPerson == "" || req.LegalIdentityCard == "" {
		// 	RespData(c, http.StatusOK, code, "商户地址信息必填")
		// 	return
		// }
		if req.BusinessUsername == "" {
			RespData(c, http.StatusOK, code, "商户注册账号id必填")
			return
		}

		if req.Branchesname == "" {
			RespData(c, http.StatusOK, code, "商户店铺名称必填")
			return
		}
		if req.ContactMobile == "" {
			RespData(c, http.StatusOK, code, "联系手机必填")
			return
		}
		if req.ImageURL == "" {
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
			code = codes.ERROR
			RespData(c, http.StatusOK, code, err.Error())
		} else {
			pc.logger.Debug("pc.service.AddStore ok")
			code = codes.SUCCESS
			RespOk(c, http.StatusOK, code)
		}

	}
}

//批量增加网点
func (pc *LianmiApisController) BatchAddStores(c *gin.Context) {
	code := codes.InvalidParams
	req := models.LotteryStoreReq{}
	if err := c.BindJSON(&req); err != nil {
		req.Offset = 0
		req.Limit = 20
	}
	pc.logger.Debug("GetLotteryStores",
		zap.String("Keyword", req.Keyword),
		zap.String("Province", req.Province),
		zap.String("City", req.City),
		zap.String("Area", req.Area),
		zap.String("Address", req.Address), //模糊查找
	)

	err := pc.service.BatchAddStores(&req)
	if err != nil {
		RespFail(c, http.StatusOK, codes.InvalidParams, "未找到网点信息")
		return
	}

	pc.logger.Debug("BatchAddStores ok")
	code = codes.SUCCESS
	RespOk(c, http.StatusOK, code)

}

//批量网点opk
func (pc *LianmiApisController) AdminDefaultOPK(c *gin.Context) {
	code := codes.InvalidParams

	err := pc.service.AdminDefaultOPK()
	if err != nil {
		RespFail(c, http.StatusOK, codes.InvalidParams, "未找到网点信息")
		return
	}

	pc.logger.Debug("BatchAddStores ok")
	code = codes.SUCCESS
	RespOk(c, http.StatusOK, code)

}
