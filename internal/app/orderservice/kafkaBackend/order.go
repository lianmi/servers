package kafkaBackend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	Global "github.com/lianmi/servers/api/proto/global"
	Msg "github.com/lianmi/servers/api/proto/msg"
	Order "github.com/lianmi/servers/api/proto/order"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/internal/common"

	"github.com/lianmi/servers/internal/pkg/models"

	"google.golang.org/protobuf/proto"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
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
		kc.logger.Info("Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
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

	var newSeq uint64

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

		//推送通知给关注的用户
		watchingUsers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("BeWatching:%s", username), "-inf", "+inf"))
		for _, watchingUser := range watchingUsers {
			//构造回包里的数据
			if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", watchingUser))); err != nil {
				kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "INCR Error"
				goto COMPLETE
			}

			//7-5 新商品上架事件 将商品信息序化
			addProductEventRsp := &Order.AddProductEventRsp{
				Username:    username,        //商户用户账号id
				Product:     req.Product,     //商品详情
				OrderType:   req.OrderType,   //订单类型，必填
				OpkBusiness: req.OpkBusiness, //商户的协商公钥，适用于任务类
				Expire:      req.Expire,      //商品过期时间
				TimeAt:      uint64(time.Now().Unix()),
			}
			productData, _ := proto.Marshal(addProductEventRsp)

			body := &Msg.MessageNotificationBody{
				Type:           Msg.MessageNotificationType_MNT_AddProduct, //关注的商户上架了新商品
				HandledAccount: username,
				HandledMsg:     "",
				Status:         1,           //消息状态
				Data:           productData, //AddProductEventRsp
				To:             watchingUser,
			}
			bodyData, _ := proto.Marshal(body)
			eRsp := &Msg.RecvMsgEventRsp{
				Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
				Type:         Msg.MessageType_MsgType_Notification, //通知类型
				Body:         bodyData,
				From:         username,                           //谁发的
				FromDeviceId: deviceID,                           //哪个设备发的
				ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
				Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
				Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
				Time:         uint64(time.Now().Unix()),
			}

			go kc.BroadcastMsgToAllDevices(eRsp, watchingUser)
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
		kc.logger.Info("Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

//7-3 商品编辑更新
func (kc *KafkaClient) HandleUpdateProduct(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64

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

		//推送通知给关注的用户
		watchingUsers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("BeWatching:%s", username), "-inf", "+inf"))
		for _, watchingUser := range watchingUsers {
			//构造回包里的数据
			if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", watchingUser))); err != nil {
				kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "INCR Error"
				goto COMPLETE
			}

			//7-6 已有商品的编辑更新事件
			updateProductEventReq := &Order.UpdateProductEventRsp{
				Username:  username,
				Product:   req.Product,
				OrderType: req.OrderType,
				Expire:    req.Expire,
				TimeAt:    uint64(time.Now().Unix()),
			}
			productData, _ := proto.Marshal(updateProductEventReq)

			body := &Msg.MessageNotificationBody{
				Type:           Msg.MessageNotificationType_MNT_UpdateProduct, //商户更新商品
				HandledAccount: username,
				HandledMsg:     "",
				Status:         1,           //消息状态
				Data:           productData, // 用来存储UpdateProductEventRsp
				To:             watchingUser,
			}
			bodyData, _ := proto.Marshal(body)
			eRsp := &Msg.RecvMsgEventRsp{
				Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
				Type:         Msg.MessageType_MsgType_Notification, //通知类型
				Body:         bodyData,
				From:         username,                           //谁发的
				FromDeviceId: deviceID,                           //哪个设备发的
				ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
				Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
				Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
				Time:         uint64(time.Now().Unix()),
			}

			go kc.BroadcastMsgToAllDevices(eRsp, watchingUser)
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

//7-4 商品下架
func (kc *KafkaClient) HandleSoldoutProduct(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleSoldoutProduct start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("SoldoutProduct",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Order.SoldoutProductReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		kc.logger.Debug("SoldoutProduct  payload",
			zap.String("ProductId", req.ProductID),
		)

		if req.ProductID == "" {
			kc.logger.Warn("下架商品id必须非空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ProductId is empty[Username=%s]", username)
			goto COMPLETE
		}

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
			kc.logger.Warn("用户不是商户类型，不能下架商品")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is not business type[Username=%s]", username)
			goto COMPLETE
		}

		//判断是否是上架
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Products:%s", username), req.ProductID); err == nil {
			if reply == nil {
				//此商品没有上架过
				kc.logger.Warn("此商品没有上架过")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Product is not onsell [Username=%s]", username)
				goto COMPLETE
			}

		}
		//TODO 判断是否存在着此商品id的订单

		//得到此商品的详细信息，如图片等，从阿里云OSS里删除这些文件
		product := new(models.Product)
		if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("Product:%s", req.ProductID))); err == nil {
			if err := redis.ScanStruct(result, product); err != nil {
				kc.logger.Error("错误：ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "HGETALL Error"
				goto COMPLETE
			}
		}
		if err = kc.DeleteAliyunOssFile(product); err != nil {
			kc.logger.Error("DeleteAliyunOssFile", zap.Error(err))
		}

		//从MySQL删除此商品
		if err = kc.DeleteProduct(req.ProductID, username); err != nil {
			kc.logger.Error("错误：从MySQL删除对应的req.ProductID失败", zap.Error(err))
		}

		//推送通知给关注的用户
		watchingUsers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("BeWatching:%s", username), "-inf", "+inf"))
		for _, watchingUser := range watchingUsers {
			//构造回包里的数据
			if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", watchingUser))); err != nil {
				kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "INCR Error"
				goto COMPLETE
			}

			//7-7 商品下架事件
			soldoutProductEventRsp := &Order.SoldoutProductEventRsp{
				ProductID: req.ProductID,
			}
			productData, _ := proto.Marshal(soldoutProductEventRsp)

			body := &Msg.MessageNotificationBody{
				Type:           Msg.MessageNotificationType_MNT_SelloutProduct, //商户下架商品
				HandledAccount: username,
				HandledMsg:     "",
				Status:         1,           //消息状态
				Data:           productData, // 用来存储SoldoutProductEventRsp
				To:             watchingUser,
			}
			bodyData, _ := proto.Marshal(body)
			eRsp := &Msg.RecvMsgEventRsp{
				Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
				Type:         Msg.MessageType_MsgType_Notification, //通知类型
				Body:         bodyData,
				From:         username,                           //谁发的
				FromDeviceId: deviceID,                           //哪个设备发的
				ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
				Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
				Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
				Time:         uint64(time.Now().Unix()),
			}

			go kc.BroadcastMsgToAllDevices(eRsp, watchingUser)
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info(" Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

//9-1 商户上传订单DH加密公钥
func (kc *KafkaClient) HandleRegisterPreKeys(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	// var newSeq uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleRegisterPreKeys start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("RegisterPreKeys",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Order.RegisterPreKeysReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		kc.logger.Debug("RegisterPreKeys  payload",
			zap.Strings("ProductId", req.GetPreKeys()),
		)

		if len(req.GetPreKeys()) == 0 {
			kc.logger.Warn("一次性公钥的数组长度必须大于0")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("PreKeys is empty[Username=%s]", username)
			goto COMPLETE
		}

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
			kc.logger.Warn("用户不是商户类型，不能上传OPK")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is not business type[Username=%s]", username)
			goto COMPLETE
		}

		//TODO opk入库
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info(" Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
向目标用户账号的所有端推送系统通知
业务号： BusinessType_Msg(5)
业务子号： MsgSubType_RecvMsgEvent(2)
*/
func (kc *KafkaClient) BroadcastMsgToAllDevices(rsp *Msg.RecvMsgEventRsp, toUser string) error {
	data, _ := proto.Marshal(rsp)

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//删除7天前的缓存系统消息
	nTime := time.Now()
	yesTime := nTime.AddDate(0, 0, -7).Unix()
	_, err := redisConn.Do("ZREMRANGEBYSCORE", fmt.Sprintf("systemMsgAt:%s", toUser), "-inf", yesTime)

	//Redis里缓存此消息,目的是用户从离线状态恢复到上线状态后同步这些系统消息给用户
	systemMsgAt := time.Now().Unix()
	if _, err := redisConn.Do("ZADD", fmt.Sprintf("systemMsgAt:%s", toUser), systemMsgAt, rsp.GetServerMsgId()); err != nil {
		kc.logger.Error("ZADD Error", zap.Error(err))
	}

	//系统消息具体内容
	key := fmt.Sprintf("systemMsg:%s:%s", toUser, rsp.GetServerMsgId())

	_, err = redisConn.Do("HMSET",
		key,
		"Username", toUser,
		"SystemMsgAt", systemMsgAt,
		"Seq", rsp.Seq,
		"Data", data,
	)

	_, err = redisConn.Do("EXPIRE", key, 7*24*3600) //设置有效期为7天

	//向toUser所有端发送
	deviceListKey := fmt.Sprintf("devices:%s", toUser)
	deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
	//查询出当前在线所有主从设备
	for _, eDeviceID := range deviceIDSliceNew {

		targetMsg := &models.Message{}
		curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
		curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
		kc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

		targetMsg.UpdateID()
		//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
		targetMsg.BuildRouter("Order", "", "Order.Frontend")

		targetMsg.SetJwtToken(curJwtToken)
		targetMsg.SetUserName(toUser)
		targetMsg.SetDeviceID(eDeviceID)
		// kickMsg.SetTaskID(uint32(taskId))
		targetMsg.SetBusinessTypeName("User")
		targetMsg.SetBusinessType(uint32(Global.BusinessType_Msg))           //消息模块
		targetMsg.SetBusinessSubType(uint32(Global.MsgSubType_RecvMsgEvent)) //接收消息事件

		targetMsg.BuildHeader("OrderService", time.Now().UnixNano()/1e6)

		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

		targetMsg.SetCode(200) //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Order.Frontend"
		if err := kc.Produce(topic, targetMsg); err == nil {
			kc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			kc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
		}

		kc.logger.Info("BroadcastMsgToAllDevices Succeed",
			zap.String("Username:", toUser),
			zap.String("DeviceID:", curDeviceKey),
			zap.Int64("Now", time.Now().Unix()))

		_ = err

	}

	return nil
}

func (kc *KafkaClient) DeleteAliyunOssFile(product *models.Product) error {
	// New client
	client, err := oss.New(common.Endpoint, common.AccessID, common.AccessKey)
	if err != nil {
		return err
	}

	// 获取存储空间。
	bucket, err := client.Bucket(common.BucketName)
	if err != nil {
		return err
	}

	//删除文件
	if product.ProductPic1 != "" {
		err = bucket.DeleteObject(product.ProductPic1)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic1:", product.ProductPic1))
		}

	}
	if product.ProductPic2 != "" {
		err = bucket.DeleteObject(product.ProductPic2)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic2:", product.ProductPic2))
		}

	}
	if product.ProductPic3 != "" {
		err = bucket.DeleteObject(product.ProductPic3)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic3:", product.ProductPic3))
		}

	}
	if product.ProductPic4 != "" {
		err = bucket.DeleteObject(product.ProductPic4)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic4:", product.ProductPic4))
		}

	}
	if product.ProductPic5 != "" {
		err = bucket.DeleteObject(product.ProductPic5)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic5:", product.ProductPic5))
		}

	}
	if product.ShortVideo1 != "" {
		err = bucket.DeleteObject(product.ShortVideo1)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ShortVideo1:", product.ShortVideo1))
		}

	}
	if product.ShortVideo2 != "" {
		err = bucket.DeleteObject(product.ShortVideo2)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ShortVideo2:", product.ShortVideo2))
		}

	}
	if product.ShortVideo3 != "" {
		err = bucket.DeleteObject(product.ShortVideo3)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ShortVideo3:", product.ShortVideo3))
		}

	}

	return nil
}
