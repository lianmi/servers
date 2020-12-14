package repositories

import (
	// "crypto/md5"
	// "encoding/hex"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	simpleJson "github.com/bitly/go-simplejson"
	"github.com/gomodule/redigo/redis"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/sts"
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

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//TODO 根据OrderID 查询  OrderImagesHistory 表的对应数据
	where := models.OrderImagesHistory{
		OrderID: req.OrderID,
	}
	orderImagesHistory := new(models.OrderImagesHistory)
	if err := s.db.Model(&models.OrderImagesHistory{}).Where(&where).First(orderImagesHistory).Error; err != nil {
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

		//超级用户
		client, err := oss.New(LMCommon.Endpoint, LMCommon.SuperAccessID, LMCommon.SuperAccessKeySecret)
		if err != nil {
			s.logger.Error("oss.New Error", zap.Error(err))
			return nil, errors.Wrapf(err, "oss失败[OrderID=%s]", req.OrderID)

		} else {
			// OSS操作。
			s.logger.Debug("创建OSSClient实例 ok")
		}
		// 获取存储空间。
		bucket, err := client.Bucket(LMCommon.BucketName)
		if err != nil {
			s.logger.Error("client.Bucket Error", zap.Error(err))
			return nil, errors.Wrapf(err, "client.Bucket失败[OrderID=%s]", req.OrderID)
		}

		//下载
		// filePath := "/tmp/" + fileName
		// err = bucket.GetObjectToFile(orderImagesHistory.BusinessOssImages, filePath)
		// if err != nil {
		// 	s.logger.Error("GetObjectToFile Error", zap.Error(err))
		// 	return nil, errors.Wrapf(err, "下载失败[OrderID=%s]", req.OrderID)
		// } else {
		// 	s.logger.Debug("下载完成: ", zap.String("filePath", filePath))
		// }

		//上传到买家oss
		// f, err := os.Open(filePath)
		// if err != nil {
		// 	s.logger.Error("os.Open Error", zap.Error(err))
		// 	return nil, errors.Wrapf(err, "os.Open失败[OrderID=%s]", req.OrderID)
		// }

		// defer f.Close()

		// md5hash := md5.New()
		// if _, err := io.Copy(md5hash, f); err != nil {
		// 	// log.Println("Copy", err)
		// 	return nil, errors.Wrapf(err, "os.Open失败[OrderID=%s]", req.OrderID)
		// }

		// md5hash.Sum(nil)
		// // log.Printf("%x\n", md5hash.Sum(nil))

		// md5Str := hex.EncodeToString(md5hash.Sum(nil))

		// var descObjectKey = strings.Replace(orderImagesHistory.BusinessOssImages, orderImagesHistory.BusinessUsername, orderImagesHistory.BuyUsername, 1)
		// _, err = bucket.CopyObject(orderImagesHistory.BusinessOssImages, descObjectKey)
		destObjectName := "orders/" + orderImagesHistory.BuyUsername + "/" + time.Now().Format("2006/01/02/") + fileName
		s.logger.Debug("复制", zap.String("destObjectName", destObjectName))

		// 拷贝文件到同一个存储空间的另一个文件。
		_, err = bucket.CopyObject(orderImagesHistory.BusinessOssImages, destObjectName)

		if err != nil {
			s.logger.Error("v Error", zap.Error(err))
			return nil, errors.Wrapf(err, "CopyObject失败[OrderID=%s]", req.OrderID)
		} else {
			s.logger.Debug("订单图片已经复制到买家oss目录", zap.String("destObjectName", destObjectName))
		}

		return &Order.DownloadOrderImagesResp{
			//订单ID
			OrderID: orderImagesHistory.OrderID,
			//商户注册id
			BusinessUsername: orderImagesHistory.BusinessUsername,
			//订单拍照图片
			Image: destObjectName,
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

func (s *MysqlLianmiRepository) RefreshOssSTSToken() error {
	var err error
	var client *sts.AliyunStsClient
	var url string

	client = sts.NewStsClient(LMCommon.AccessID, LMCommon.AccessKey, LMCommon.RoleAcs)
	//生成阿里云oss临时sts, Policy是对lianmi-ipfs这个bucket下的 avatars, generalavatars, msg, products, stores, teamicons, 目录有可读写权限

	// Policy是对lianmi-ipfs这个bucket下的user目录有可读写权限
	acsAvatars := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/avatars/*")
	acsGeneralavatars := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/generalavatars/*")
	acsMsg := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/msg/*")
	acsProducts := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/products/*")
	acsStores := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/stores/*")
	acsOrders := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/orders/*")
	acsTeamIcons := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/teamicons/*")
	acsUsers := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/users/*")

	// Policy是对lianmi-ipfs这个bucket下的user目录有可读写权限
	policy := sts.Policy{
		Version: "1",
		Statement: []sts.StatementBase{sts.StatementBase{
			Effect:   "Allow",
			Action:   []string{"oss:GetObject", "oss:ListObjects", "oss:PutObject", "oss:AbortMultipartUpload"},
			Resource: []string{acsAvatars, acsGeneralavatars, acsMsg, acsProducts, acsStores, acsOrders, acsTeamIcons, acsUsers},
		}},
	}

	//1小时过期
	url, err = client.GenerateSignatureUrl("lianmiserver", fmt.Sprintf("%d", LMCommon.EXPIRESECONDS), policy.ToJson())
	if err != nil {
		s.logger.Error("GenerateSignatureUrl Error", zap.Error(err))
		return err
	}

	data, err := client.GetStsResponse(url)
	if err != nil {
		s.logger.Error("阿里云oss GetStsResponse Error", zap.Error(err))
		return err
	}

	// log.Println("result:", string(data))
	sjson, err := simpleJson.NewJson(data)
	if err != nil {
		s.logger.Warn("simplejson.NewJson Error", zap.Error(err))
		return err
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
		return err

	}

	//保存到redis里
	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	redisConn.Do("SET", "OSSAccessKeyId", accessKeyID)
	redisConn.Do("EXPIRE", "OSSAccessKeyId", LMCommon.EXPIRESECONDS) //设置失效时间为1小时

	redisConn.Do("SET", "OSSAccessKeySecret", accessSecretKey)
	redisConn.Do("EXPIRE", "OSSAccessKeySecret", LMCommon.EXPIRESECONDS) //设置失效时间为1小时

	redisConn.Do("SET", "OSSSecurityToken", securityToken)
	redisConn.Do("EXPIRE", "OSSSecurityToken", LMCommon.EXPIRESECONDS) //设置失效时间为1小时

	return nil
}
