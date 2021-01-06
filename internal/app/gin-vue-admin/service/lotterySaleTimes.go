package service

import (
	"github.com/lianmi/servers/internal/app/gin-vue-admin/global"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/model/request"
	"github.com/lianmi/servers/internal/pkg/models"
)

//@author: [piexlmax](https://github.com/piexlmax)
//@function: CreateLotterySaleTimes
//@description: 创建LotterySaleTimes记录
//@param: lotterySaleTimes models.LotterySaleTime
//@return: err error

func CreateLotterySaleTimes(lotterySaleTimes models.LotterySaleTime) (err error) {
	err = global.GVA_DB.Create(&lotterySaleTimes).Error
	return err
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: DeleteLotterySaleTimes
//@description: 删除LotterySaleTimes记录
//@param: lotterySaleTimes models.LotterySaleTime
//@return: err error

func DeleteLotterySaleTimes(lotterySaleTimes models.LotterySaleTime) (err error) {
	err = global.GVA_DB.Delete(&lotterySaleTimes).Error
	return err
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: DeleteLotterySaleTimesByIds
//@description: 批量删除LotterySaleTimes记录
//@param: ids request.IdsReq
//@return: err error

func DeleteLotterySaleTimesByIds(ids request.IdsReq) (err error) {
	err = global.GVA_DB.Delete(&[]models.LotterySaleTime{}, "id in ?", ids.Ids).Error
	return err
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: UpdateLotterySaleTimes
//@description: 更新LotterySaleTimes记录
//@param: lotterySaleTimes *models.LotterySaleTime
//@return: err error

func UpdateLotterySaleTimes(lotterySaleTimes models.LotterySaleTime) (err error) {
	err = global.GVA_DB.Save(&lotterySaleTimes).Error
	return err
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: GetLotterySaleTimes
//@description: 根据id获取LotterySaleTimes记录
//@param: id uint
//@return: err error, lotterySaleTimes models.LotterySaleTime

func GetLotterySaleTimes(id uint) (err error, lotterySaleTimes models.LotterySaleTime) {
	err = global.GVA_DB.Where("id = ?", id).First(&lotterySaleTimes).Error
	return
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: GetLotterySaleTimesInfoList
//@description: 分页获取LotterySaleTimes记录
//@param: info request.LotterySaleTimesSearch
//@return: err error, list interface{}, total int64

func GetLotterySaleTimesInfoList(info request.LotterySaleTimesSearch) (err error, list interface{}, total int64) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := global.GVA_DB.Model(&models.LotterySaleTime{})
	var lotterySaleTimess []models.LotterySaleTime
	// 如果有条件搜索 下方会自动创建搜索语句
	err = db.Count(&total).Error
	err = db.Limit(limit).Offset(offset).Find(&lotterySaleTimess).Error
	return err, lotterySaleTimess, total
}
