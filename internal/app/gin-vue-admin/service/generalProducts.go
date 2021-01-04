package service

import (
	"github.com/lianmi/servers/internal/app/gin-vue-admin/global"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/model"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/model/request"
)

//@author: [piexlmax](https://github.com/piexlmax)
//@function: CreateGeneralProduct
//@description: 创建GeneralProduct记录
//@param: generalProducts model.GeneralProduct
//@return: err error

func CreateGeneralProduct(generalProducts model.GeneralProduct) (err error) {
	err = global.GVA_DB.Create(&generalProducts).Error
	return err
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: DeleteGeneralProduct
//@description: 删除GeneralProduct记录
//@param: generalProducts model.GeneralProduct
//@return: err error

func DeleteGeneralProduct(generalProducts model.GeneralProduct) (err error) {
	err = global.GVA_DB.Delete(&generalProducts).Error
	return err
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: DeleteGeneralProductByIds
//@description: 批量删除GeneralProduct记录
//@param: ids request.IdsReq
//@return: err error

func DeleteGeneralProductByIds(ids request.IdsReq) (err error) {
	err = global.GVA_DB.Delete(&[]model.GeneralProduct{},"id in ?",ids.Ids).Error
	return err
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: UpdateGeneralProduct
//@description: 更新GeneralProduct记录
//@param: generalProducts *model.GeneralProduct
//@return: err error

func UpdateGeneralProduct(generalProducts model.GeneralProduct) (err error) {
	err = global.GVA_DB.Save(&generalProducts).Error
	return err
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: GetGeneralProduct
//@description: 根据id获取GeneralProduct记录
//@param: id uint
//@return: err error, generalProducts model.GeneralProduct

func GetGeneralProduct(id uint) (err error, generalProducts model.GeneralProduct) {
	err = global.GVA_DB.Where("id = ?", id).First(&generalProducts).Error
	return
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: GetGeneralProductInfoList
//@description: 分页获取GeneralProduct记录
//@param: info request.GeneralProductSearch
//@return: err error, list interface{}, total int64

func GetGeneralProductInfoList(info request.GeneralProductSearch) (err error, list interface{}, total int64) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&model.GeneralProduct{})
    var generalProductss []model.GeneralProduct
    // 如果有条件搜索 下方会自动创建搜索语句
	err = db.Count(&total).Error
	err = db.Limit(limit).Offset(offset).Find(&generalProductss).Error
	return err, generalProductss, total
}