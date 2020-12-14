package repositories

import (
	// "crypto/md5"
	// "encoding/hex"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	// simpleJson "github.com/bitly/go-simplejson"
	"github.com/gomodule/redigo/redis"
	// LMCommon "github.com/lianmi/servers/internal/common"
	// "github.com/lianmi/servers/internal/pkg/sts"s
	// "io"
	// "os"
	// "path"
	"path/filepath"

	// Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"

	"github.com/lianmi/servers/internal/pkg/models"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	// "strings"
	"time"
)

var (
	endpoint = "https://oss-cn-hangzhou.aliyuncs.com"
	// // 阿里云主账号AccessKey拥有所有API的访问权限，风险很高。强烈建议您创建并使用RAM账号进行API访问或日常运维，请登录 https://ram.console.aliyun.com 创建RAM账号。
	accessID        = "LTAI4FzZsweRdNRd3KLsUc2J"
	accessKeySecret = "W8a576pxtoyiJ7n8g4RHBFz9k5fF3r"
	bucketName      = "lianmi-ipfs"

	client *oss.Client
	bucket *oss.Bucket
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
	var err error
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//TODO 根据OrderID 查询  OrderImagesHistory 表的对应数据
	where := models.OrderImagesHistory{
		OrderID: req.OrderID,
	}
	orderImagesHistory := new(models.OrderImagesHistory)
	if err = s.db.Model(&models.OrderImagesHistory{}).Where(&where).First(orderImagesHistory).Error; err != nil {
		// if errors.Is(err, gorm.ErrRecordNotFound) {}
		return nil, errors.Wrapf(err, "Record is not exists[OrderID=%s]", req.OrderID)
	} else {

		//文件名
		fileName := filepath.Base(orderImagesHistory.BusinessOssImages)

		s.logger.Debug("DownloadOrderImages",
			zap.String("OrderID", orderImagesHistory.OrderID),
			zap.String("ProductID", orderImagesHistory.ProductID),
			zap.String("BuyUsername", orderImagesHistory.BuyUsername),
			zap.String("BusinessUsername", orderImagesHistory.BusinessUsername),
			zap.String("BuyerOssImages", orderImagesHistory.BuyerOssImages),
			zap.String("BusinessOssImages", orderImagesHistory.BusinessOssImages),
			zap.String("fileName", fileName),
			zap.Float64("Cost", orderImagesHistory.Cost),
			zap.Uint64("BlockNumber", orderImagesHistory.BlockNumber),
			zap.String("TxHash", orderImagesHistory.TxHash),
		)

		// 超级用户创建OSSClient实例。
		client, err = oss.New(endpoint, accessID, accessKeySecret)

		if err != nil {
			return nil, errors.Wrapf(err, "oss.New失败[OrderID=%s]", req.OrderID)

		}

		// 获取存储空间。
		bucket, err = client.Bucket(bucketName)
		if err != nil {
			return nil, errors.Wrapf(err, "client.Bucket失败[OrderID=%s]", req.OrderID)

		}

		signedURL, err := bucket.SignURL(orderImagesHistory.BuyUsername, oss.HTTPGet, 60)
		if err != nil {
			s.logger.Error("bucket.SignURL error", zap.Error(err))
			return nil, errors.Wrapf(err, "bucket.SignURL失败[OrderID=%s]", req.OrderID)
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
