package repositories

import (
	"fmt"
	
	"github.com/gomodule/redigo/redis"
	// Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"

	"github.com/lianmi/servers/internal/pkg/models"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	// "strings"
	// "time"
)

func (s *MysqlLianmiRepository) GetOrderInfo(orderID string) (*models.OrderInfo, error) {
	var err error
	var curState int
	var productID string
	var buyUser, businessUser string
	var attachHash string
	var orderTotalAmount float64 //订单金额
	var isPayed bool

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//订单详情
	orderIDKey := fmt.Sprintf("Order:%s", orderID)
	//获取订单的具体信息
	curState, err = redis.Int(redisConn.Do("HGET", orderIDKey, "State"))
	isPayed, err = redis.Bool(redisConn.Do("HGET", orderIDKey, "IsPayed"))
	productID, err = redis.String(redisConn.Do("HGET", orderIDKey, "ProductID"))
	buyUser, err = redis.String(redisConn.Do("HGET", orderIDKey, "BuyUser"))
	businessUser, err = redis.String(redisConn.Do("HGET", orderIDKey, "BusinessUser"))
	orderTotalAmount, err = redis.Float64(redisConn.Do("HGET", orderIDKey, "OrderTotalAmount"))
	attachHash, err = redis.String(redisConn.Do("HGET", orderIDKey, "AttachHash"))
	if err != nil {
		return nil, err
	}
	return &models.OrderInfo{
		OrderID:          orderID,
		AttachHash:       attachHash,
		ProductID:        productID,
		BuyerUsername:    buyUser,
		BusinessUsername: businessUser,
		Cost:             orderTotalAmount,
		State:            curState,
		IsPayed:          isPayed,
	}, nil

}

//增加订单拍照图片上链历史表
func (s *MysqlLianmiRepository) SaveOrderImagesBlockchain(req *Order.UploadOrderImagesReq, orderTotalAmount float64, blcokNumber uint64, buyUser, businessUser, hash string) error {
	//TODO 将字段增加到 OrderImagesHistory 表，然后将订单图片复制到买家id的orders目录
	//先查询OrderID的数据是否存在，如果存在，则返回，如果不存在，则新增
	where := models.OrderImagesHistory{
		OrderID: req.OrderID,
	}
	orderImagesHistory := new(models.OrderImagesHistory)
	if err := s.db.Model(&models.OrderImagesHistory{}).Where(&where).First(orderImagesHistory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			oi := &models.OrderImagesHistory{
				OrderID:          req.OrderID,
				BuyUsername:      buyUser,
				BusinessUsername: businessUser,
				Cost:             orderTotalAmount,
				// BuyerOssImages:    descObjectKey,
				BusinessOssImages: req.Image,
				BlockNumber:       blcokNumber,
				TxHash:            hash,
			}
			if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&oi).Error; err != nil {
				s.logger.Error("UploadOrderImages, failed to upsert OrderImages History", zap.Error(err))
				return err
			} else {
				s.logger.Debug("UploadOrderImages, upsert OrderImages History succeed")
			}
			return nil

		} else {
			return err
		}

	} else {
		return errors.Wrapf(err, "Record is exists[OrderID=%s]", req.OrderID)
	}

}

//用户端: 根据 OrderID 获取OrderImages表对应的所有订单拍照图片
func (s *MysqlLianmiRepository) DownloadOrderImages(req *Order.DownloadOrderImagesReq) (*Order.DownloadOrderImagesResp, error) {
	//TODO 根据OrderID 查询  OrderImagesHistory 表的对应数据
	where := models.OrderImagesHistory{
		OrderID: req.OrderID,
	}
	orderImagesHistory := new(models.OrderImagesHistory)
	if err := s.db.Model(&models.OrderImagesHistory{}).Where(&where).First(orderImagesHistory).Error; err != nil {
		// if errors.Is(err, gorm.ErrRecordNotFound) {}
		return nil, errors.Wrapf(err, "Record is not exists[OrderID=%s]", req.OrderID)
	} else {
		s.logger.Debug("DownloadOrderImages",
			zap.String("OrderID", orderImagesHistory.OrderID),
			zap.String("ProductID", orderImagesHistory.ProductID),
			zap.String("BuyUsername", orderImagesHistory.BuyUsername),
			zap.String("BusinessUsername", orderImagesHistory.BusinessUsername),
			zap.String("BuyerOssImages", orderImagesHistory.BuyerOssImages),
			zap.String("BusinessOssImages", orderImagesHistory.BusinessOssImages),
			zap.Float64("Cost", orderImagesHistory.Cost),
			zap.Uint64("BlockNumber", orderImagesHistory.BlockNumber),
			zap.String("TxHash", orderImagesHistory.TxHash),
		)

		//将商户的订单图片下载到临时目录/tmp/，然后上传到orders买家的对应目录


	}
	return nil, nil
}
