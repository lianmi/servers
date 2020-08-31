package kafkaBackend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	Order "github.com/lianmi/servers/api/proto/order"
	User "github.com/lianmi/servers/api/proto/user"

	"github.com/lianmi/servers/internal/pkg/models"

	"google.golang.org/protobuf/proto"

	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

/*
商户的商品表在redis里用有序集合保存， Products:{username}, 分数为更新时间， 元素是productID
商品详细表用哈希表ProductInfo:{productID}保存

*/

/*
1. 根据timeAt增量返回商品信息，首次timeAt请初始化为0，服务器返回全量商品信息，后续采取增量方式更新
2. 如果soldoutProducts不为空，终端根据soldoutProducts移除商品缓存数据
3. 获取商品信息的流程： 发起获取商品信息请求 → 更新本地数据库 → 返回数据给UI

*/
func (kc *KafkaClient) HandleQueryProducts(msg *models.Message) error {
	var err error
	// var toUser, teamID string
	errorCode := 200
	var errorMsg string
	rsp := &Order.QueryProductsRsp{
		Products:        make([]*Order.Product, 0),
		SoldoutProducts: make([]string, 0),
		TimeAt:          uint64(time.Now().Unix()),
	}

	// var newSeq uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleQueryProducts start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("QueryProducts",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Order.QueryProductsReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		kc.logger.Debug("QueryProducts  payload",
			zap.String("UserName", req.GetUserName()),
			zap.Uint64("TimeAt", req.GetTimeAt()),
		)
		//从redis里获取当前用户信息
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				kc.logger.Error("错误：ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
				goto COMPLETE

			}
		}

		//判断此商户是不是用户关注的，如果不是则返回
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Watching:%s", req.GetUserName()), username); err == nil {
			if reply == nil {
				//商户不是用户关注
				kc.logger.Debug("商户不是用户关注",
					zap.String("UserName", req.GetUserName()),
				)
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("User is not watching[Username=%s]", username)
				goto COMPLETE
			}

		}

		//获取商户的商品有序集合
		//从redis的有序集合查询出商户的商品信息在时间戳req.GetTimeAt()之后的更新
		productIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Products:%s", req.GetUserName()), req.GetTimeAt(), "+inf"))
		for _, productID := range productIDs {
			product := new(models.Product)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("Product:%s", productID))); err == nil {
				if err := redis.ScanStruct(result, product); err != nil {
					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					continue
				}
			}

			rsp.Products = append(rsp.Products, &Order.Product{
				ProductId:         productID,
				ProductName:       product.ProductName,
				CategoryName:      product.CategoryName,
				ProductDesc:       product.ProductDesc,
				ProductPic1:       product.ProductPic1,
				ProductPic2:       product.ProductPic2,
				ProductPic3:       product.ProductPic3,
				ProductPic4:       product.ProductPic4,
				ProductPic5:       product.ProductPic5,
				ShortVideo1:       product.ShortVideo1,
				ShortVideo2:       product.ShortVideo2,
				ShortVideo3:       product.ShortVideo3,
				Price:             product.Price,
				LeftCount:         product.LeftCount,
				Discount:          product.Discount,
				DiscountDesc:      product.DiscountDesc,
				DiscountStartTime: uint64(product.DiscountStartTime),
				DiscountEndTime:   uint64(product.DiscountEndTime),
				CreateAt:          uint64(product.CreateAt),
				ModifyAt:          uint64(product.ModifyAt),
			})
		}

		//获取商户的下架soldoutProducts
		soldoutProductIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("SoldoutProducts:%s", req.GetUserName()), req.GetTimeAt(), "+inf"))
		for _, soldoutProductID := range soldoutProductIDs {
			rsp.SoldoutProducts = append(rsp.SoldoutProducts, soldoutProductID)
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ := proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("SendMsgRsp message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send SendMsgRsp message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

//7-2 商品上架
func (kc *KafkaClient) HandleAddProduct(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	rsp := &Order.AddProductRsp{}

	// var newSeq uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleAddProduct start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("AddProduct",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Order.AddProductReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		kc.logger.Debug("AddProduct  payload",
			zap.String("ProductId", req.GetProduct().ProductId),
			zap.Int("OrderType", int(req.GetOrderType())),
			zap.String("OpkBusiness", req.GetOpkBusiness()),
			zap.Uint64("Expire", req.GetExpire()),
		)

		if req.GetProduct().ProductId != "" {
			kc.logger.Warn("新的上架商品id必须是空的 ")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ProductId is not empty[Username=%s]", username)
			goto COMPLETE
		}

		//生成随机的商品id
		req.Product.ProductId = uuid.NewV4().String()
		rsp.ProductID = req.Product.ProductId

		//从redis里获取当前用户信息
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				kc.logger.Error("错误：ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
				goto COMPLETE

			}
		}

		if userData.UserType != int(User.UserType_Ut_Business) {
			kc.logger.Warn("用户不是商户类型，不能上架商品")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is not business type[Username=%s]", username)
			goto COMPLETE
		}

		//上架
		if _, err = redisConn.Do("ZADD", fmt.Sprintf("Products:%s", username), time.Now().Unix(), req.Product.ProductId); err != nil {
			kc.logger.Error("ZADD Error", zap.Error(err))
		}

		product := &models.Product{
			Username:          username,
			ProductID:         req.Product.ProductId,
			ProductName:       req.Product.ProductName,
			CategoryName:      req.Product.CategoryName,
			ProductDesc:       req.Product.ProductDesc,
			ProductPic1:       req.Product.ProductPic1,
			ProductPic2:       req.Product.ProductPic2,
			ProductPic3:       req.Product.ProductPic3,
			ProductPic4:       req.Product.ProductPic4,
			ProductPic5:       req.Product.ProductPic5,
			ShortVideo1:       req.Product.ShortVideo1,
			ShortVideo2:       req.Product.ShortVideo2,
			ShortVideo3:       req.Product.ShortVideo3,
			Price:             req.Product.Price,
			LeftCount:         req.Product.LeftCount,
			Discount:          req.Product.Discount,
			DiscountDesc:      req.Product.DiscountDesc,
			DiscountStartTime: int64(req.Product.DiscountStartTime),
			DiscountEndTime:   int64(req.Product.DiscountEndTime),
			CreateAt:          time.Now().Unix(),
		}

		if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("Product:%s", req.Product.ProductId)).AddFlat(product)...); err != nil {
			kc.logger.Error("错误：HMSET TeamInfo", zap.Error(err))
		}

		//保存到MySQL
		if err = kc.SaveProduct(product); err != nil {
			kc.logger.Error("错误：保存到MySQL失败", zap.Error(err))
		}

		//TODO 推送通知给关注的用户
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ := proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("SendMsgRsp message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send SendMsgRsp message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

//7-3 商品编辑更新
func (kc *KafkaClient) HandleUpdateProduct(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	// rsp := &Order.AddProductRsp{}

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleUpdateProduct start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("UpdateProduct",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Order.UpdateProductReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		kc.logger.Debug("UpdateProduct  payload",
			zap.String("ProductId", req.GetProduct().ProductId),
			zap.Int("OrderType", int(req.GetOrderType())),
			zap.Uint64("Expire", req.GetExpire()),
		)

		if req.GetProduct().ProductId == "" {
			kc.logger.Warn("上架商品id必须非空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ProductId is empty[Username=%s]", username)
			goto COMPLETE
		}

		// req.Product.ProductId

		//从redis里获取当前用户信息
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				kc.logger.Error("错误：ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
				goto COMPLETE

			}
		}

		if userData.UserType != int(User.UserType_Ut_Business) {
			kc.logger.Warn("用户不是商户类型，不能上架商品")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is not business type[Username=%s]", username)
			goto COMPLETE
		}

		//判断是否是上架
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Products:%s", username), req.Product.ProductId); err == nil {
			if reply == nil {
				//此商品没有上架过
				kc.logger.Warn("此商品没有上架过")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Product is not onsell [Username=%s]", username)
				goto COMPLETE
			}

		}

		product := &models.Product{
			Username:          username,
			ProductID:         req.Product.ProductId,
			ProductName:       req.Product.ProductName,
			CategoryName:      req.Product.CategoryName,
			ProductDesc:       req.Product.ProductDesc,
			ProductPic1:       req.Product.ProductPic1,
			ProductPic2:       req.Product.ProductPic2,
			ProductPic3:       req.Product.ProductPic3,
			ProductPic4:       req.Product.ProductPic4,
			ProductPic5:       req.Product.ProductPic5,
			ShortVideo1:       req.Product.ShortVideo1,
			ShortVideo2:       req.Product.ShortVideo2,
			ShortVideo3:       req.Product.ShortVideo3,
			Price:             req.Product.Price,
			LeftCount:         req.Product.LeftCount,
			Discount:          req.Product.Discount,
			DiscountDesc:      req.Product.DiscountDesc,
			DiscountStartTime: int64(req.Product.DiscountStartTime),
			DiscountEndTime:   int64(req.Product.DiscountEndTime),
			CreateAt:          time.Now().Unix(),
		}

		if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("Product:%s", req.Product.ProductId)).AddFlat(product)...); err != nil {
			kc.logger.Error("错误：HMSET TeamInfo", zap.Error(err))
		}

		//保存到MySQL
		if err = kc.SaveProduct(product); err != nil {
			kc.logger.Error("错误：保存到MySQL失败", zap.Error(err))
		}

		//TODO 推送通知给关注的用户
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		// data, _ := proto.Marshal(rsp)
		// msg.FillBody(data)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("SendMsgRsp message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send SendMsgRsp message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}
