package repositories

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	simpleJson "github.com/bitly/go-simplejson"
	"github.com/gomodule/redigo/redis"
	// Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/internal/pkg/sts"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	// "time"
)

//商户端: 将完成订单拍照所有图片上链
func (s *MysqlLianmiRepository) UploadOrderImages(req *Order.UploadOrderImagesReq) error {
	//TODO 将字段增加到 OrderImagesHistory 表，然后将订单图片复制到买家id的orders目录
	//先查询OrderID的数据是否存在，如果存在，则返回，如果不存在，则新增
	where := models.OrderImagesHistory{
		OrderID: req.OrderID,
	}
	orderImagesHistory := new(models.OrderImagesHistory)
	if err := s.db.Model(&models.OrderImagesHistory{}).Where(&where).First(orderImagesHistory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var curState int
			var productID string
			var buyUser, businessUser string
			var attachHash string
			var orderTotalAmount float64 //订单金额
			var isPayed bool

			redisConn := s.redisPool.Get()
			defer redisConn.Close()

			//订单详情
			orderIDKey := fmt.Sprintf("Order:%s", req.OrderID)
			//获取订单的具体信息
			curState, err = redis.Int(redisConn.Do("HGET", orderIDKey, "State"))
			isPayed, err = redis.Bool(redisConn.Do("HGET", orderIDKey, "IsPayed"))
			productID, err = redis.String(redisConn.Do("HGET", orderIDKey, "ProductID"))
			buyUser, err = redis.String(redisConn.Do("HGET", orderIDKey, "BuyUser"))
			businessUser, err = redis.String(redisConn.Do("HGET", orderIDKey, "BusinessUser"))
			orderTotalAmount, err = redis.Float64(redisConn.Do("HGET", orderIDKey, "OrderTotalAmount"))
			attachHash, err = redis.String(redisConn.Do("HGET", orderIDKey, "AttachHash"))

			if err != nil {
				s.logger.Error("从Redis里取出此 Order 对应的businessUser Error", zap.String("orderIDKey", orderIDKey), zap.Error(err))
			}

			/*
				暂时屏蔽， 不判断支付是否成功
				if !isPayed {
					s.logger.Error("Order is not Payed")

					return errors.Wrapf(err, "Order is not Payed[OrderID=%s]", req.OrderID)
				}
			*/
			_ = isPayed

			if productID == "" {
				s.logger.Error("ProductID is empty")

				return errors.Wrapf(err, "ProductID is empty[OrderID=%s]", req.OrderID)
			}

			if buyUser == "" {
				s.logger.Error("BuyUser is empty")
				return errors.Wrapf(err, "BuyUser is empty[OrderID=%s]", req.OrderID)
			}

			if businessUser == "" {
				s.logger.Error("BusinessUser is empty")
				return errors.Wrapf(err, "BusinessUser is empty[OrderID=%s]", req.OrderID)
			}

			s.logger.Debug("UploadOrderImages",
				zap.Int("State", curState), //状态
				zap.String("OrderID", req.OrderID),
				zap.String("ProductID", productID),
				zap.String("BuyUser", buyUser),
				zap.String("BusinessUser", businessUser),
				zap.String("AttachHash", attachHash), //订单内容hash
				zap.Float64("OrderTotalAmount", orderTotalAmount),
				zap.String("OrderImageFile", req.Image),
			)
			/*
				 暂时屏蔽，不成功
					// 将订单图片复制到买家的orders私有目录
					var client *sts.AliyunStsClient
					var url string

					client = sts.NewStsClient(LMCommon.AccessID, LMCommon.AccessKey, LMCommon.RoleAcs)
					//生成阿里云oss临时sts, Policy是对lianmi-ipfs这个bucket下的 avatars, generalavatars, msg, products, stores, teamicons, 目录有可读写权限

					// Policy是对lianmi-ipfs这个bucket下的user目录有可读写权限
					policy := sts.Policy{
						Version: "1",
						Statement: []sts.StatementBase{sts.StatementBase{
							Effect:   "Allow",
							Action:   []string{"oss:*"},
							Resource: []string{"acs:oss:*:*:lianmi-ipfs/orders/*"},
						}},
					}

					url, err = client.GenerateSignatureUrl("client", fmt.Sprintf("%d", LMCommon.EXPIRESECONDS), policy.ToJson())
					if err != nil {
						s.logger.Error("GenerateSignatureUrl Error", zap.Error(err))
						return errors.Wrapf(err, "Oss error[OrderID=%s]", req.OrderID)
					}

					data, err := client.GetStsResponse(url)
					if err != nil {
						s.logger.Error("阿里云oss GetStsResponse Error", zap.Error(err))
						return errors.Wrapf(err, "Oss error[OrderID=%s]", req.OrderID)
					}

					// log.Println("result:", string(data))
					sjson, err := simpleJson.NewJson(data)
					if err != nil {
						s.logger.Warn("simplejson.NewJson Error", zap.Error(err))
						return errors.Wrapf(err, "NewJson error[OrderID=%s]", req.OrderID)
					}
					accessKeyID := sjson.Get("Credentials").Get("AccessKeyId").MustString()
					accessSecretKey := sjson.Get("Credentials").Get("AccessKeySecret").MustString()
					securityToken := sjson.Get("Credentials").Get("SecurityToken").MustString()

					s.logger.Debug("收到阿里云OSS服务端的回包",
						zap.String("RequestId", sjson.Get("RequestId").MustString()),
						zap.String("AccessKeyId", accessKeyID),
						zap.String("AccessKeySecret", accessSecretKey),
						zap.String("SecurityToken", securityToken),
						zap.String("Expiration", sjson.Get("Credentials").Get("Expiration").MustString()),
					)

					if accessKeyID == "" || accessSecretKey == "" || securityToken == "" {
						s.logger.Warn("获取STS错误")
						return errors.Wrapf(err, "Oss error[OrderID=%s]", req.OrderID)
					}
					// Copy an existing object
					// 创建OSSClient实例。
					client2, err := oss.New(LMCommon.Endpoint, accessKeyID, accessSecretKey, oss.SecurityToken(securityToken))
					if err != nil {
						s.logger.Error("阿里云oss Error", zap.Error(err))
						return errors.Wrapf(err, "Oss error[OrderID=%s]", req.OrderID)

					} else {
						// OSS操作。
						s.logger.Debug("利用临时STS创建OSSClient实例 ok")
					}

					// 获取存储空间
					bucket, err := client2.Bucket(LMCommon.BucketName)
					if err != nil {
						s.logger.Error("阿里云oss Error, client2.Bucket", zap.Error(err))
						return errors.Wrapf(err, "Oss error[OrderID=%s]", req.OrderID)
					}

					var descObjectKey = strings.Replace(req.Image, businessUser, buyUser, 1)
					s.logger.Debug("After Replace ", zap.String("descObjectKey", descObjectKey))

					_, err = bucket.CopyObject(req.Image, descObjectKey)
					if err != nil {
						s.logger.Error("阿里云oss Error, bucket.CopyObject", zap.Error(err))
						return errors.Wrapf(err, "Oss error[OrderID=%s]", req.OrderID)
					} else {
						s.logger.Debug("CopyObject ok", zap.String("req.Image", req.Image), zap.String("descObjectKey", descObjectKey))
					}
					//{"level":"error","ts":1607868562.76748,"msg":"阿里云oss Error, bucket.CopyObject","type":"LianmiRepository","error":"Put \"https://lianmi-ipfs.oss-cn-hangzhou.aliyuncs.com/orders%2Fid58%2F2020%2F12%2F13%2F73bc66f54d22094a633a617f09391cf7.jpeg\": x509: certificate signed by unknown authority"}
			*/
			oi := &models.OrderImagesHistory{
				OrderID:           req.OrderID,
				BuyUsername:       buyUser,
				BussinessUsername: businessUser,
				Cost:              orderTotalAmount,
				// BuyerOssImages:    descObjectKey,
				BusinessOssImages: req.Image,
				BlockNumber:       uint64(0),
				TxHash:            "",
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
	//TODO
	return nil, nil
}
