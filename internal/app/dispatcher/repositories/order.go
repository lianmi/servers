package repositories

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gomodule/redigo/redis"
	LMCommon "github.com/lianmi/servers/internal/common"

	Order "github.com/lianmi/servers/api/proto/order"

	"github.com/lianmi/servers/internal/pkg/models"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func (s *MysqlLianmiRepository) GetOrderInfo(orderID string) (*models.OrderInfo, error) {
	// var err error
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
	curState, _ = redis.Int(redisConn.Do("HGET", orderIDKey, "State"))
	isPayed, _ = redis.Bool(redisConn.Do("HGET", orderIDKey, "IsPayed"))
	productID, _ = redis.String(redisConn.Do("HGET", orderIDKey, "ProductID"))
	buyUser, _ = redis.String(redisConn.Do("HGET", orderIDKey, "BuyUser"))
	businessUser, _ = redis.String(redisConn.Do("HGET", orderIDKey, "BusinessUser"))
	orderTotalAmount, _ = redis.Float64(redisConn.Do("HGET", orderIDKey, "OrderTotalAmount"))
	attachHash, _ = redis.String(redisConn.Do("HGET", orderIDKey, "AttachHash"))
	// if err != nil {
	// 	s.logger.Error("UploadOrderImages, HGET Error", zap.Error(err))
	// 	return nil, err
	// }
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
				BusinessOssImage: req.Image,
				BlockNumber:      blcokNumber,
				TxHash:           hash,
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
func (s *MysqlLianmiRepository) DownloadOrderImage(orderID string) (*Order.DownloadOrderImagesResp, error) {
	var err error
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//根据OrderID 查询  OrderImagesHistory 表的对应数据
	where := models.OrderImagesHistory{
		OrderID: orderID,
	}
	orderImagesHistory := new(models.OrderImagesHistory)
	if err = s.db.Model(&models.OrderImagesHistory{}).Where(&where).First(orderImagesHistory).Error; err != nil {
		// if errors.Is(err, gorm.ErrRecordNotFound) {}
		return nil, errors.Wrapf(err, "Record is not exists[OrderID=%s]", orderID)
	} else {

		s.logger.Debug("DownloadOrderImage",
			zap.String("OrderID", orderImagesHistory.OrderID),
			zap.String("ProductID", orderImagesHistory.ProductID),
			zap.String("BuyUsername", orderImagesHistory.BuyUsername),
			zap.String("BusinessUsername", orderImagesHistory.BusinessUsername),
			zap.String("BusinessOssImage", orderImagesHistory.BusinessOssImage),
			zap.Float64("Cost", orderImagesHistory.Cost),
			zap.Uint64("BlockNumber", orderImagesHistory.BlockNumber),
			zap.String("TxHash", orderImagesHistory.TxHash),
		)

		// 超级用户创建OSSClient实例。
		client, err := oss.New(LMCommon.Endpoint, LMCommon.SuperAccessID, LMCommon.AccessKey)

		if err != nil {
			return nil, errors.Wrapf(err, "oss.New失败[OrderID=%s]", orderID)

		}

		// 获取存储空间。
		bucket, err := client.Bucket(LMCommon.BucketName)
		if err != nil {
			return nil, errors.Wrapf(err, "client.Bucket失败[OrderID=%s]", orderID)

		}

		//生成签名URL下载链接， 300s后过期
		signedURL, err := bucket.SignURL(orderImagesHistory.BusinessOssImage, oss.HTTPGet, 300)
		if err != nil {
			s.logger.Error("bucket.SignURL error", zap.Error(err))
			return nil, errors.Wrapf(err, "bucket.SignURL失败[OrderID=%s]", orderID)
		}

		return &Order.DownloadOrderImagesResp{
			//订单ID
			OrderID: orderImagesHistory.OrderID,
			//商户注册id
			BusinessUsername: orderImagesHistory.BusinessUsername,
			//订单拍照图片URL
			ImageURL: signedURL,
			// 区块高度
			BlockNumber: orderImagesHistory.BlockNumber,
			// 交易哈希hex
			Hash: orderImagesHistory.TxHash,
			//时间
			Time: uint64(time.Now().UnixNano() / 1e6),
		}, nil

	}
	return nil, nil
}
