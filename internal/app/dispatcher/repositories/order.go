package repositories

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gomodule/redigo/redis"

	// "github.com/lianmi/servers/api/proto/global"
	LMCommon "github.com/lianmi/servers/internal/common"

	"github.com/lianmi/servers/api/proto/global"
	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"

	"github.com/lianmi/servers/internal/pkg/models"

	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *MysqlLianmiRepository) GetOrderInfo(orderID string) (*models.OrderInfo, error) {
	// var err error
	var curState int
	var bodyType int
	var productID string
	var buyUser, businessUser string
	var attach, attachHash string
	var opkBuyUser, opkBusinessUser string
	var bodyObjFile string
	var orderImageFile string
	var orderTotalAmount float64 //订单金额
	var isPayed bool
	var blockNumber uint64
	var ticketCode uint64
	var txHash string

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//订单详情
	orderIDKey := fmt.Sprintf("Order:%s", orderID)
	//获取订单的具体信息
	ticketCode, _ = redis.Uint64(redisConn.Do("HGET", orderIDKey, "TicketCode"))
	curState, _ = redis.Int(redisConn.Do("HGET", orderIDKey, "State"))
	isPayed, _ = redis.Bool(redisConn.Do("HGET", orderIDKey, "IsPayed"))
	productID, _ = redis.String(redisConn.Do("HGET", orderIDKey, "ProductID"))
	buyUser, _ = redis.String(redisConn.Do("HGET", orderIDKey, "BuyUser"))
	businessUser, _ = redis.String(redisConn.Do("HGET", orderIDKey, "BusinessUser"))
	orderTotalAmount, _ = redis.Float64(redisConn.Do("HGET", orderIDKey, "OrderTotalAmount"))
	attach, _ = redis.String(redisConn.Do("HGET", orderIDKey, "attacch"))
	attachHash, _ = redis.String(redisConn.Do("HGET", orderIDKey, "AttachHash"))
	bodyType, _ = redis.Int(redisConn.Do("HGET", orderIDKey, "BodyType"))
	bodyObjFile, _ = redis.String(redisConn.Do("HGET", orderIDKey, "BodyObjFile"))
	orderImageFile, _ = redis.String(redisConn.Do("HGET", orderIDKey, "OrderImageFile"))
	blockNumber, _ = redis.Uint64(redisConn.Do("HGET", orderIDKey, "BlockNumber"))

	txHash, _ = redis.String(redisConn.Do("HGET", orderIDKey, "TxHash"))
	opkBuyUser, _ = redis.String(redisConn.Do("HGET", orderIDKey, "OpkBuyUser"))
	opkBusinessUser, _ = redis.String(redisConn.Do("HGET", orderIDKey, "OpkBusinessUser"))

	return &models.OrderInfo{
		OrderID:          orderID,
		TicketCode:       ticketCode,
		ProductID:        productID,
		Attach:           attach,
		AttachHash:       attachHash,
		BodyType:         bodyType,
		BodyObjFile:      bodyObjFile,
		OrderImageFile:   orderImageFile,
		BuyerUsername:    buyUser,
		OpkBuyUser:       opkBuyUser,
		BusinessUsername: businessUser,
		OpkBusinessUser:  opkBusinessUser,
		Cost:             orderTotalAmount,
		State:            curState,
		IsPayed:          isPayed,
		BlockNumber:      blockNumber,
		TxHash:           txHash,
	}, nil

}

//增加订单拍照图片上链历史表
func (s *MysqlLianmiRepository) SaveOrderImagesBlockchain(req *Order.UploadOrderImagesReq, orderTotalAmount float64, blcokNumber uint64, buyUser, businessUser, hash string) error {
	//保存到redis及 MySQL
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//订单详情
	orderIDKey := fmt.Sprintf("Order:%s", req.OrderID)
	redisConn.Do("HSET", orderIDKey, "OrderImageFile", req.Image)

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

//修改订单的body类型及body加密阿里云文件上链历史表
func (s *MysqlLianmiRepository) SaveOrderBody(req *Order.UploadOrderBodyReq) error {
	var err error
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//保存到redis里
	orderIDKey := fmt.Sprintf("Order:%s", req.OrderID)
	_, err = redisConn.Do("HSET", orderIDKey, "BodyType", int(req.BodyType))
	_, err = redisConn.Do("HSET", orderIDKey, "BodyObjFile", req.BodyObjFile)

	where := models.OrderImagesHistory{
		OrderID: req.OrderID,
	}

	// 同时更新多个字段
	result := s.db.Model(&models.OrderImagesHistory{}).Where(&where).Updates(&models.OrderImagesHistory{
		BodyType:    int(req.BodyType),
		BodyObjFile: req.BodyObjFile,
	})

	//updated records count
	s.logger.Debug("SaveOrderBody result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("修改订单的body类型及body加密阿里云文件上链历史表失败", zap.Error(result.Error))
		return result.Error
	} else {
		s.logger.Debug("修改订单的body类型及body加密阿里云文件上链历史表成功")
	}
	_ = err
	return nil
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
		s.logger.Warn("DownloadOrderImage 无法找到订单id对应的图片", zap.String("OrderID", orderImagesHistory.OrderID))
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
		client, err := oss.New(LMCommon.Endpoint, LMCommon.SuperAccessID, LMCommon.SuperAccessKeySecret)

		if err != nil {
			return nil, errors.Wrapf(err, "oss.New失败[OrderID=%s]", orderID)

		}

		// 获取存储空间。
		bucket, err := client.Bucket(LMCommon.BucketName)
		if err != nil {
			return nil, errors.Wrapf(err, "client.Bucket失败[OrderID=%s]", orderID)

		}

		//生成签名URL下载链接， 300s后过期
		objectName := orderImagesHistory.BusinessOssImage
		// objectName = "orders/id58/2021/01/16/6E05AD9D654ADFAD155901843E71B870"

		signedURL, err := bucket.SignURL(objectName, oss.HTTPGet, 300)
		if err != nil {
			s.logger.Error("bucket.SignURL error", zap.Error(err))
			return nil, errors.Wrapf(err, "bucket.SignURL失败[OrderID=%s]", orderID)
		} else {
			s.logger.Debug("bucket.SignURL 生成成功")

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

}

//创建新订单
//redis缓存订单各种信息，当订单的状态变化时，需要更新redis
func (s *MysqlLianmiRepository) SavaOrderItemToDB(item *models.OrderItems) error {
	//panic("implement me")
	var err error
	var ticketCode uint64
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	currentTime := time.Now()
	item.CreatedAt = currentTime.UnixNano() / 1e6
	//将订单ID保存到商户的订单有序集合orders:{username}，订单详情是 orderInfo:{订单ID}
	if _, err := redisConn.Do("ZADD", fmt.Sprintf("orders:%s", item.UserId), item.CreatedAt, item.OrderId); err != nil {
		s.logger.Error("ZADD Error", zap.Error(err))
		return err
	}

	//每个商户都有自己的出票码并递增
	if ticketCode, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("TicketCode:%s", item.StoreId))); err != nil {
		s.logger.Error("redisConn INCR TicketCode Error", zap.Error(err))

		return err
	}

	//订单详情
	_, err = redisConn.Do("HMSET",
		fmt.Sprintf("Order:%s", item.OrderId),
		"OrderType", int(Global.OrderType_ORT_Normal), //订单类型是普通订单
		"BuyUser", item.UserId, //发起订单的用户id
		"OpkBuyUser", item.PublicKey,
		"BusinessUser", item.StoreId, //商户的用户id
		"OpkBusinessUser", item.StorePublicKey,
		"OrderID", item.OrderId, //订单id
		"TicketCode", ticketCode, //出票码
		"ProductID", item.ProductId, //商品id
		// "Type", req.OrderType, //订单类型
		"State", item.OrderStatus, //订单状态,
		"Attach", item.Body, //订单内容
		"AttachHash", "", //订单内容attach的哈希值， 默认为空
		"IsPayed", LMCommon.REDISFALSE, //此订单支付状态， true- 支付完成，false-未支付
		"IsUrge", LMCommon.REDISFALSE, //催单
		"CreateAt", uint64(item.CreatedAt), //毫秒
	)
	if err != nil {
		s.logger.Error("HMSET Error", zap.Error(err))
		return err
	}

	item.TicketCode = int64(ticketCode) //出票码
	err = s.db.Create(item).Error
	return err
}

//多条件分页查询某个用户或商户的订单列表，status =0 返回所有用户
func (s *MysqlLianmiRepository) GetOrderListByUser(username string, limit int, offset, status int) (p *[]models.OrderItems, err error) {
	//panic("implement me")
	p = new([]models.OrderItems)
	if status == 0 {
		err = s.db.Model(&models.OrderItems{}).Where(&models.OrderItems{UserId: username}).Or(&models.OrderItems{StoreId: username}).Limit(limit).Offset(offset).Find(p).Error
	} else {
		// err = s.db.Model(&models.OrderItems{}).Where(" ( user_id = ? or store_id = ? ) and order_status = ?  ", username, username, status).Limit(limit).Offset(offset).Find(p).Error

		columns := []string{"*"}
		orderBy := "updated_at desc"

		redisConn := s.redisPool.Get()
		defer redisConn.Close()

		wheres := make([]interface{}, 0)
		wheres = append(wheres, []interface{}{"order_status", status})

		if username != "" {
			wheres = append(wheres, []interface{}{"user_id = ? or store_id = ?", username, username})
		}

		db2 := s.db
		db2, err = s.base.BuildQueryList(db2, wheres, columns, orderBy, offset, limit)
		if err != nil {
			return nil, err
		}
		err = db2.Find(p).Error

		if err != nil {
			s.logger.Error("Find错误", zap.Error(err))
			return nil, err
		}

	}
	return
}

func (s *MysqlLianmiRepository) GetOrderListByID(orderID string) (p *models.OrderItems, err error) {
	//panic("implement me")
	p = new(models.OrderItems)
	err = s.db.Model(&models.OrderItems{}).Where(&models.OrderItems{OrderId: orderID}).First(p).Error
	return
}

func (s *MysqlLianmiRepository) SetOrderStatusByOrderID(orderID string, status int) error {
	//panic("implement me")
	err := s.db.Model(&models.OrderItems{}).Where(&models.OrderItems{OrderId: orderID}).Updates(&models.OrderItems{OrderStatus: status}).Error
	return err
}

func CheckOrderStatusUpdataRules(currentStatus, newStatus int) bool {
	//  完成的状态 , 和取消的 不可以修改
	//if currentStatus == int(global.OrderState_OS_Done) || currentStatus == int(global.OrderState_OS_Cancel){
	//	return false
	//}
	//
	//// 初始化的可以修改任意状态
	//if currentStatus == int(global.OrderState_OS_Undefined) {
	//	return true
	//}
	//
	//// TODO 支付后的不能修改为 特定的场景 等
	//if currentStatus == int(global.OrderState_OS_IsPayed) && newStatus == int(global.OrderState_OS_Undefined)||
	// currentStatus == int(global.OrderState_OS_IsPayed) && newStatus == int(global.OrderState_OS_Paying){
	//	return false
	//}

	// 白名单模式
	// 从未支付到支付完成

	// 各种过滤条件
	isokInitToPayingOrPayedOrApplyCancel := currentStatus == int(global.OrderState_OS_Undefined) && (newStatus == int(global.OrderState_OS_IsPayed) || newStatus == int(global.OrderState_OS_Paying) || newStatus == int(global.OrderState_OS_ApplyCancel))
	isokApplyCancelToCancel := currentStatus == int(global.OrderState_OS_ApplyCancel) && newStatus == int(global.OrderState_OS_Cancel)
	isokPayingToPayed := currentStatus == int(global.OrderState_OS_Paying) && newStatus == int(global.OrderState_OS_IsPayed)
	isokIsPayedOrProcessingToProcessingOrTaked := (currentStatus == int(global.OrderState_OS_IsPayed) || currentStatus == int(global.OrderState_OS_Processing)) && (newStatus == int(global.OrderState_OS_Processing) || newStatus == int(global.OrderState_OS_Taked) || newStatus == int(global.OrderState_OS_Refuse))
	isokTakedToDone := newStatus == int(global.OrderState_OS_Taked) || newStatus == int(global.OrderState_OS_Done)
	if isokInitToPayingOrPayedOrApplyCancel ||
		isokApplyCancelToCancel ||
		isokPayingToPayed ||
		isokIsPayedOrProcessingToProcessingOrTaked ||
		isokTakedToDone {
		return true
	}

	return false

}

// 修改订单状态接口
// 仅能处理 拒单,接单,确认收获这三种状态
// 其他状态均不可以想这个接口处理
func (s *MysqlLianmiRepository) UpdateOrderStatus(userid string, storeID string, orderid string, newStatus int) (p *models.OrderItems, err error) {
	//panic("implement me")
	// 获取当前的订单信息
	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	// 用户有效性 入口已经处理

	// 100k * 1000 = 10m

	// 在redis 获取订单的状态
	currentOrderStatus, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("Order:%s", orderid), "State"))
	if err != nil {
		// 订单信息异常
		s.logger.Error("订单信息异常", zap.Error(err))
		return nil, fmt.Errorf("订单信息异常")
	}

	// 订单状态修改的方向有
	// 用户 : 发送订单 -> 取消订单 -- 不支付, 手动可以取消订单
	// 系统支付回调 : 发送订单 -> 支付完成 -- 支付回调 触发退款任务
	// 商户: 支付完成 -> 拒单 -- 服务端触发退款
	// 商户: 支付完成 -> 已接单 -- 通过上传彩票照片接口触发
	// 用户: 已接单 -> 确认收货 -- 这时候 会推送到 见证中心
	// 见证中心回调: 确认收货 -> 上联成功 -- 回调接口
	//

	// 总结 : 这个接口只处理 以下几个状态的变化 , 其他都拒绝处理
	// 商户: 支付完成 -> 拒单 -- 服务端触发退款
	// 商户: 支付完成 -> 已接单 -- 通过上传彩票照片接口触发
	// 用户: 已接单 -> 确认收货 -- 这时候 会推送到 见证中心

	// 判断当前状态 是不是商户 从 已支付 -> 拒单
	isokBusinessPayedToRefuse := (Global.OrderState(currentOrderStatus) == Global.OrderState_OS_IsPayed) && (Global.OrderState(newStatus) == Global.OrderState_OS_Refuse)
	// 判断当前状态 是不是商户 从 已支付 -> 完成订单
	isokBusinessPayedToDone := (Global.OrderState(currentOrderStatus) == Global.OrderState_OS_IsPayed) && (Global.OrderState(newStatus) == Global.OrderState_OS_Done)
	// 判断当前状态是不是 用户从 完成订单 -> 确认收货
	isokUserDoneToConfirm := (Global.OrderState(currentOrderStatus) == Global.OrderState_OS_Done) && (Global.OrderState(newStatus) == Global.OrderState_OS_Confirm)

	if isokBusinessPayedToDone ||
		isokBusinessPayedToRefuse ||
		isokUserDoneToConfirm {
		// 满足条件的状态 才通过

	} else {
		s.logger.Error("用户无权修改的状态方向 ", zap.Int("开始状态", currentOrderStatus), zap.Int("结束状态", newStatus))
		return nil, fmt.Errorf("用户无权操作这个状态的变化")
	}
	// 以下三种状态相互独立不会同时出现

	// 没有其他的可以处理
	//// 商户 从 已支付 -> 拒单
	//if isokBusinessPayedToRefuse {
	//
	//}
	////商户 从 已支付 -> 完成订单
	//if isokBusinessPayedToDone {
	//
	//}
	//
	//// 用户从 完成订单 -> 确认收货
	//if isokUserDoneToConfirm {
	//
	//}

	_, err = redisConn.Do("HMSET",
		fmt.Sprintf("Order:%s", orderid),

		// "Type", req.OrderType, //订单类型

		"State", newStatus, //订单状态,初始为 发送订单

	)
	if err != nil {
		s.logger.Error("HMSET Error", zap.Error(err))
		return nil, fmt.Errorf("修改状态失败")
	}
	// 成功同时更新到数据库

	p = new(models.OrderItems)
	errFind := s.db.Model(p).Where(&models.OrderItems{UserId: userid, StoreId: storeID, OrderId: orderid}).First(p).Error
	if errFind != nil {
		s.logger.Error("UpdateOrderStatus DB ", zap.Error(errFind))
		return nil, fmt.Errorf("未找到订单信息")
	}

	// 订单状态判断
	// 更新状态
	err = s.db.Model(&models.OrderItems{}).Where(&models.OrderItems{UserId: userid, StoreId: storeID, OrderId: orderid}).Updates(&models.OrderItems{OrderStatus: newStatus}).Error
	p.OrderStatus = newStatus
	return
}

//根据微信支付回调更改订单状态
func (s *MysqlLianmiRepository) UpdateOrderStatusByWechatCallback(orderid string) error {
	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	// 用户有效性 入口已经处理

	// 100k * 1000 = 10m

	// 在redis 获取订单的状态
	currentOrderStatus, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("Order:%s", orderid), "State"))
	if err != nil {
		// 订单信息异常
		s.logger.Error("订单信息异常", zap.Error(err))
		return fmt.Errorf("订单信息异常")
	}

	//如果当前状态不是支付中，则不允许修改
	if currentOrderStatus != int(global.OrderState_OS_Paying) {
		return fmt.Errorf("可更换状态错误")
	}

	_, err = redisConn.Do("HMSET",
		fmt.Sprintf("Order:%s", orderid),

		// "Type", req.OrderType, //订单类型
		"State", int(Global.OrderState_OS_IsPayed), //订单状态,初始为 发送订单

	)
	if err != nil {
		s.logger.Error("HMSET Error", zap.Error(err))
		return fmt.Errorf("修改状态失败")
	}
	// 成功同时更新到数据库

	p := new(models.OrderItems)
	errFind := s.db.Model(p).Where(&models.OrderItems{OrderId: orderid}).First(p).Error
	if errFind != nil {
		s.logger.Error("UpdateOrderStatus DB ", zap.Error(errFind))
		return fmt.Errorf("未找到订单信息")
	}

	// 订单状态判断
	// 更新状态
	err = s.db.Model(&models.OrderItems{}).Where(&models.OrderItems{OrderId: orderid}).Updates(&models.OrderItems{OrderStatus: int(global.OrderState_OS_IsPayed)}).Error

	return err

}

func (s *MysqlLianmiRepository) GetStoreOpkByBusiness(id string) (string, error) {
	//panic("implement me")
	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	opk, _ := redis.String(redisConn.Do("GET", fmt.Sprintf("DefaultOPK:%s", id)))

	if opk == "" {
		return "", fmt.Errorf("商户opk找不到")
	} else {
		s.logger.Debug("GetStoreOpkByBusiness", zap.String("商户id", id), zap.String("商户协商公钥", opk))
		return opk, nil
	}
}

func (s *MysqlLianmiRepository) OrderPushPrize(username string, orderID string, prize float64) (string, error) {
	//panic("implement me")
	// 获取订单信息
	//BusinessUser
	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	// 用户有效性 入口已经处理

	// 100k * 1000 = 10m

	// 在redis 获取订单的状态
	currentOrderStatus, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("Order:%s", orderID), "State"))
	if err != nil {
		// 订单信息异常
		s.logger.Error("订单信息异常", zap.Error(err))
		return "", fmt.Errorf("订单信息异常")
	}

	// 判断订单状态是否可以更改

	if currentOrderStatus != int(global.OrderState_OS_Confirm) {
		return "", fmt.Errorf("订单状态未确认 , 无法兑奖")
	}

	currentOrderStoreID, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("Order:%s", orderID), "BusinessUser"))
	if err != nil {
		// 订单信息异常
		s.logger.Error("订单信息异常", zap.Error(err))
		return "", fmt.Errorf("订单信息异常")
	}

	if currentOrderStoreID != username {
		return "", fmt.Errorf("当前订单的商户不是当前操作的商户", zap.String("订单的商户", currentOrderStoreID), zap.String("操作的商户", username))
	}

	currentOrderBuyUserID, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("Order:%s", orderID), "BuyUser"))
	if err != nil {
		// 订单信息异常
		s.logger.Error("订单信息异常", zap.Error(err))
		return "", fmt.Errorf("订单信息异常")
	}

	// 更新 订单信息

	_, err = redisConn.Do("HMSET",
		fmt.Sprintf("Order:%s", orderID),
		// "Type", req.OrderType, //订单类型
		"State", int(Global.OrderState_OS_Prizeed), //订单状态,初始为 发送订单
		"Prize", prize, // 兑奖的金额
	)
	if err != nil {
		s.logger.Error("HMSET Error", zap.Error(err))
		return "", fmt.Errorf("修改状态失败")
	}

	// 更新数据库
	err = s.db.Model(&models.OrderItems{}).Where(&models.OrderItems{OrderId: orderID, StoreId: username, OrderStatus: int(global.OrderState_OS_Prizeed)}).Updates(&models.OrderItems{Prize: prize}).Error

	return currentOrderBuyUserID, err
}
