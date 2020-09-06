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
3. 获取商品信息的流程:  发起获取商品信息请求 → 更新本地数据库 → 返回数据给UI

*/
func (kc *KafkaClient) HandleQueryProducts(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	rsp := &Order.QueryProductsRsp{
		Products:        make([]*Order.Product, 0),
		SoldoutProducts: make([]string, 0),
		TimeAt:          uint64(time.Now().UnixNano() / 1e6),
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

				kc.logger.Error("错误: ScanStruct", zap.Error(err))
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
					kc.logger.Error("错误: ScanStruct", zap.Error(err))
					continue
				}
			}

			rsp.Products = append(rsp.Products, &Order.Product{
				ProductId:         productID,
				Expire:            uint64(product.Expire),
				ProductName:       product.ProductName,
				CategoryName:      product.CategoryName,
				ProductDesc:       product.ProductDesc,
				ProductPic1Small:  product.ProductPic1Small,
				ProductPic1Middle: product.ProductPic1Middle,
				ProductPic1Large:  product.ProductPic1Large,

				ProductPic2Small:  product.ProductPic2Small,
				ProductPic2Middle: product.ProductPic2Middle,
				ProductPic2Large:  product.ProductPic2Large,

				ProductPic3Small:  product.ProductPic3Small,
				ProductPic3Middle: product.ProductPic3Middle,
				ProductPic3Large:  product.ProductPic3Large,

				Thumbnail:         product.Thumbnail,
				ShortVideo:        product.ShortVideo,
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
			zap.String("OpkBusinessUser", req.GetOpkBusinessUser()),
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

				kc.logger.Error("错误: ScanStruct", zap.Error(err))
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
		if _, err = redisConn.Do("ZADD", fmt.Sprintf("Products:%s", username), time.Now().UnixNano()/1e6, req.Product.ProductId); err != nil {
			kc.logger.Error("ZADD Error", zap.Error(err))
		}

		product := &models.Product{
			Username:          username,
			ProductID:         req.Product.ProductId,
			Expire:            int64(req.Product.Expire),
			ProductName:       req.Product.ProductName,
			CategoryName:      req.Product.CategoryName,
			ProductDesc:       req.Product.ProductDesc,
			ProductPic1Small:  req.Product.ProductPic1Small,
			ProductPic1Middle: req.Product.ProductPic1Middle,
			ProductPic1Large:  req.Product.ProductPic1Large,

			ProductPic2Small:  req.Product.ProductPic2Small,
			ProductPic2Middle: req.Product.ProductPic2Middle,
			ProductPic2Large:  req.Product.ProductPic2Large,

			ProductPic3Small:  req.Product.ProductPic3Small,
			ProductPic3Middle: req.Product.ProductPic3Middle,
			ProductPic3Large:  req.Product.ProductPic3Large,

			Thumbnail:         req.Product.Thumbnail,
			ShortVideo:        req.Product.ShortVideo,
			Price:             req.Product.Price,
			LeftCount:         req.Product.LeftCount,
			Discount:          req.Product.Discount,
			DiscountDesc:      req.Product.DiscountDesc,
			DiscountStartTime: int64(req.Product.DiscountStartTime),
			DiscountEndTime:   int64(req.Product.DiscountEndTime),
			CreateAt:          time.Now().UnixNano() / 1e6,
		}

		if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("Product:%s", req.Product.ProductId)).AddFlat(product)...); err != nil {
			kc.logger.Error("错误: HMSET ProductInfo", zap.Error(err))
		}

		//保存到MySQL
		if err = kc.SaveProduct(product); err != nil {
			kc.logger.Error("错误: 保存到MySQL失败", zap.Error(err))
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
				Username:    username,            //商户用户账号id
				Product:     req.Product,         //商品详情
				OrderType:   req.OrderType,       //订单类型，必填
				OpkBusiness: req.OpkBusinessUser, //商户的协商公钥，适用于任务类
				Expire:      req.Expire,          //商品过期时间
				TimeAt:      uint64(time.Now().UnixNano() / 1e6),
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
				Time:         uint64(time.Now().UnixNano() / 1e6),
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
			// zap.Uint64("Expire", req.GetExpire()),
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

				kc.logger.Error("错误: ScanStruct", zap.Error(err))
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
			Expire:            int64(req.Product.Expire),
			ProductName:       req.Product.ProductName,
			CategoryName:      req.Product.CategoryName,
			ProductDesc:       req.Product.ProductDesc,
			ProductPic1Small:  req.Product.ProductPic1Small,
			ProductPic1Middle: req.Product.ProductPic1Middle,
			ProductPic1Large:  req.Product.ProductPic1Large,

			ProductPic2Small:  req.Product.ProductPic2Small,
			ProductPic2Middle: req.Product.ProductPic2Middle,
			ProductPic2Large:  req.Product.ProductPic2Large,

			ProductPic3Small:  req.Product.ProductPic3Small,
			ProductPic3Middle: req.Product.ProductPic3Middle,
			ProductPic3Large:  req.Product.ProductPic3Large,

			Thumbnail:         req.Product.Thumbnail,
			ShortVideo:        req.Product.ShortVideo,
			Price:             req.Product.Price,
			LeftCount:         req.Product.LeftCount,
			Discount:          req.Product.Discount,
			DiscountDesc:      req.Product.DiscountDesc,
			DiscountStartTime: int64(req.Product.DiscountStartTime),
			DiscountEndTime:   int64(req.Product.DiscountEndTime),
			CreateAt:          time.Now().UnixNano() / 1e6,
		}

		if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("Product:%s", req.Product.ProductId)).AddFlat(product)...); err != nil {
			kc.logger.Error("错误: HMSET Product Info", zap.Error(err))
		}

		//保存到MySQL
		if err = kc.SaveProduct(product); err != nil {
			kc.logger.Error("错误: 保存到MySQL失败", zap.Error(err))
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
				// Expire:    req.Expire,
				TimeAt: uint64(time.Now().UnixNano() / 1e6),
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
				Time:         uint64(time.Now().UnixNano() / 1e6),
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

				kc.logger.Error("错误: ScanStruct", zap.Error(err))
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
				kc.logger.Error("错误: ScanStruct", zap.Error(err))
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
			kc.logger.Error("错误: 从MySQL删除对应的req.ProductID失败", zap.Error(err))
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
				Time:         uint64(time.Now().UnixNano() / 1e6),
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
			zap.Strings("PreKeys", req.GetPreKeys()),
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

				kc.logger.Error("错误: ScanStruct", zap.Error(err))
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

		//opk入库
		prekeys := make([]*models.Prekey, 0)
		for _, opk := range req.GetPreKeys() {
			prekeys = append(prekeys, &models.Prekey{
				Type:         0,
				Username:     username,
				Publickey:    opk,
				UploadTimeAt: time.Now().UnixNano() / 1e6,
			})

			//保存到redis里prekeys:{username}
			if _, err := redisConn.Do("ZADD", fmt.Sprintf("prekeys:%s", username), time.Now().UnixNano()/1e6, opk); err != nil {
				kc.logger.Error("ZADD Error", zap.Error(err))
			}
		}

		if err = kc.SavePreKeys(prekeys); err != nil {
			kc.logger.Error("SavePreKeys错误", zap.Error(err))
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

//9-2 获取网点OPK公钥及订单ID
func (kc *KafkaClient) HandleGetPreKeyOrderID(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	rsp := &Order.GetPreKeyOrderIDRsp{}

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleGetPreKeyOrderID start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("GetPreKeyOrderID",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Order.GetPreKeyOrderIDReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		kc.logger.Debug("GetPreKeyOrderID  payload",
			zap.String("UserName", req.GetUserName()),     //商户
			zap.Int("OrderType", int(req.GetOrderType())), //订单类型
			zap.String("ProducctID", req.GetProductID()),  //商品id
		)

		if req.GetUserName() == "" {
			kc.logger.Warn("商户用户账号不能为空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("UserName is empty[Username=%s]", req.GetUserName())
			goto COMPLETE
		}
		if req.GetProductID() == "" {
			kc.logger.Warn("商品id不能为空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ProductID is empty[ProductID=%s]", req.GetProductID())
			goto COMPLETE
		}

		//从redis里获取当前用户信息
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				kc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
				goto COMPLETE

			}
		}

		//从redis里获取目标商户的信息
		businessUserData := new(models.User)
		userKey = fmt.Sprintf("userData:%s", req.GetUserName())
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, businessUserData); err != nil {

				kc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", req.GetUserName())
				goto COMPLETE

			}
		}

		//判断商户是否被封号
		if businessUserData.State == 2 {
			kc.logger.Warn("此商户已被封号", zap.String("businessUser", req.GetUserName()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is blocked[Username=%s]", req.GetUserName())
			goto COMPLETE
		}

		if businessUserData.UserType != int(User.UserType_Ut_Normal) {
			kc.logger.Warn("目标用户不是商户类型")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is not business type[Username=%s]", req.GetUserName())
			goto COMPLETE
		}

		// 获取ProductID对应的商品信息
		product := new(models.Product)
		if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("Product:%s", req.GetProductID()))); err == nil {
			if err := redis.ScanStruct(result, product); err != nil {
				kc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("This Product is not exists")
				goto COMPLETE
			}
		}

		//检测商品有效期是否过期， 对彩票竞猜类的商品，有效期内才能下单
		if (product.Expire > 0) && (product.Expire < time.Now().UnixNano()/1e6) {
			kc.logger.Warn("商品有效期过期", zap.Int64("Expire", product.Expire))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Product is expire")
			goto COMPLETE
		}

		// 生成订单ID
		orderID := uuid.NewV4().String()

		opk := ""

		//从商户的prekeys有序集合取出一个opk
		prekeySlice, _ := redis.Strings(redisConn.Do("ZRANGE", fmt.Sprintf("prekeys:%s", req.GetUserName()), 1, 1))
		if len(prekeySlice) > 0 {
			opk = prekeySlice[0]

			//取出后就删除此OPK
			if _, err = redisConn.Do("ZREM", fmt.Sprintf("prekeys:%s", req.GetUserName()), opk); err != nil {
				kc.logger.Error("ZREM Error", zap.Error(err))
			}

		} else {
			kc.logger.Warn("商户的prekeys有序集合无法取出")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Business opks is empty[Username=%s]", req.GetUserName())
			goto COMPLETE
		}

		rsp.UserName = req.GetUserName()
		rsp.OrderType = req.GetOrderType()
		rsp.ProductID = req.GetProductID()
		rsp.OrderID = orderID
		rsp.PubKey = opk

		//将订单ID保存到商户的订单有序集合orders:{username}，订单详情是 orderInfo:{订单ID}
		if _, err := redisConn.Do("ZADD", fmt.Sprintf("orders:%s", req.GetUserName()), time.Now().UnixNano()/1e6, orderID); err != nil {
			kc.logger.Error("ZADD Error", zap.Error(err))
		}

		//订单详情
		_, err = redisConn.Do("HMSET",
			fmt.Sprintf("orderInfo:%s", orderID),
			"sourceUser", username, //发起订单的用户id
			"deviceid", deviceID, //发起订单的用户设备id
			"businessUser", req.GetUserName(), //商户的用户id
			"productID", req.GetProductID(), //商品id，默认是空
			"orderType", req.GetOrderType(), //订单类型
			"orderState", int(Global.OrderState_OS_Undefined), //订单状态,初始为0
			"createAt", uint64(time.Now().UnixNano()/1e6), //秒
		)
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
		kc.logger.Info(" Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
向目标用户账号的所有端推送消息， 接收端会触发接收消息事件
业务号:  BusinessType_Msg(5)
业务子号:  MsgSubType_RecvMsgEvent(2)
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
	systemMsgAt := time.Now().UnixNano() / 1e6
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
		targetMsg.SetBusinessTypeName("Order")
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
			zap.Int64("Now", time.Now().UnixNano()/1e6))

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
	if product.ProductPic1Small != "" {
		err = bucket.DeleteObject(product.ProductPic1Small)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic1Small:", product.ProductPic1Small))
		}

	}
	if product.ProductPic1Middle != "" {
		err = bucket.DeleteObject(product.ProductPic1Middle)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic1Middle:", product.ProductPic1Middle))
		}

	}
	if product.ProductPic1Large != "" {
		err = bucket.DeleteObject(product.ProductPic1Large)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic1Large:", product.ProductPic1Large))
		}

	}

	if product.ProductPic2Small != "" {
		err = bucket.DeleteObject(product.ProductPic2Small)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic2Small:", product.ProductPic2Small))
		}

	}
	if product.ProductPic2Middle != "" {
		err = bucket.DeleteObject(product.ProductPic2Middle)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic2Middle:", product.ProductPic2Middle))
		}

	}
	if product.ProductPic2Large != "" {
		err = bucket.DeleteObject(product.ProductPic2Large)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic2Large:", product.ProductPic2Large))
		}

	}

	if product.ProductPic3Small != "" {
		err = bucket.DeleteObject(product.ProductPic3Small)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic3Small:", product.ProductPic3Small))
		}

	}
	if product.ProductPic3Middle != "" {
		err = bucket.DeleteObject(product.ProductPic3Middle)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic3Middle:", product.ProductPic3Middle))
		}

	}
	if product.ProductPic3Large != "" {
		err = bucket.DeleteObject(product.ProductPic3Large)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic3Large:", product.ProductPic3Large))
		}

	}

	if product.Thumbnail != "" {
		err = bucket.DeleteObject(product.Thumbnail)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("Thumbnail:", product.Thumbnail))
		}

	}
	if product.ShortVideo != "" {
		err = bucket.DeleteObject(product.ShortVideo)
		if err == nil {
			kc.logger.Info("删除文件 Succeed",
				zap.String("ShortVideo:", product.ShortVideo))
		}

	}

	return nil
}

/*
9-3 下单 处理订单消息，是由ChatService转发过来的
*/
func (kc *KafkaClient) HandleOrderMsg(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64

	rsp := &Msg.SendMsgRsp{}

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleOrderMsg start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("OrderMsg",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Msg.SendMsgReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		kc.logger.Debug("OrderMsg payload",
			zap.Int32("Scene", int32(req.GetScene())),
			zap.Int32("Type", int32(req.GetType())),
			zap.String("To", req.GetTo()),
			zap.String("Uuid", req.GetUuid()),
			zap.Uint64("SendAt", req.GetSendAt()),
		)

		if req.GetTo() == "" {
			kc.logger.Warn("商户用户账号不能为空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("To is empty[Username=%s]", req.GetTo())
			goto COMPLETE
		}
		if req.GetType() != Msg.MessageType_MsgType_Order {
			kc.logger.Warn("警告，不能处理非订单类型的消息")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Type is not right[Type=%d]", int32(req.GetType()))
			goto COMPLETE
		}

		//从redis里获取当前用户信息
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				kc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
				goto COMPLETE

			}
		}

		//从redis里获取目标商户的信息
		businessUserData := new(models.User)
		userKey = fmt.Sprintf("userData:%s", req.GetTo())
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, businessUserData); err != nil {

				kc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", req.GetTo())
				goto COMPLETE

			}
		}

		//判断商户是否被封号
		if businessUserData.State == 2 {
			kc.logger.Warn("此商户已被封号", zap.String("businessUser", req.GetTo()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is blocked[Username=%s]", req.GetTo())
			goto COMPLETE
		}

		// if businessUserData.UserType != int(User.UserType_Ut_Normal) {
		// 	kc.logger.Warn("目标用户不是商户类型")
		// 	errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		// 	errorMsg = fmt.Sprintf("User is not business type[Username=%s]", req.GetTo())
		// 	goto COMPLETE
		// }
		//解包出 OrderProductBody
		var orderProductBody = new(Order.OrderProductBody)
		if err := proto.Unmarshal(req.GetBody(), orderProductBody); err != nil {
			kc.logger.Error("Protobuf Unmarshal OrderProductBody Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Protobuf Unmarshal OrderProductBody Error: %s", err.Error())
			goto COMPLETE

		} else {
			kc.logger.Debug("OrderProductBody payload",
				zap.String("OrderID", orderProductBody.GetOrderID()),
				zap.String("ProductID", orderProductBody.GetProductID()),
				zap.String("BuyUser", orderProductBody.GetBuyUser()),
				zap.String("OpkBuyUser", orderProductBody.GetOpkBuyUser()),
				zap.String("BusinessUser", orderProductBody.GetBusinessUser()),
				zap.String("OpkBusinessUser", orderProductBody.GetOpkBusinessUser()),
				zap.String("Attach", orderProductBody.GetAttach()),
			)

			if orderProductBody.GetOrderID() == "" {
				kc.logger.Error("OrderID is empty")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("OrderID is empty")
				goto COMPLETE
			}

			if orderProductBody.GetProductID() == "" {
				kc.logger.Error("ProductID is empty")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ProductID is empty")
				goto COMPLETE
			}

			if orderProductBody.GetBuyUser() == "" {
				kc.logger.Error("BuyUser is empty")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("BuyUser is empty")
				goto COMPLETE
			}

			if orderProductBody.GetOpkBuyUser() == "" {
				kc.logger.Error("OpkBuyUser is empty")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("OpkBuyUser is empty")
				goto COMPLETE
			}

			if orderProductBody.GetBusinessUser() == "" {
				kc.logger.Error("BusinessUse is empty")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("BusinessUse is empty")
				goto COMPLETE
			}

			if orderProductBody.GetOpkBusinessUser() == "" {
				kc.logger.Error("OpkBusinessUser is empty")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("OpkBusinessUser is empty")
				goto COMPLETE
			}

			// 获取ProductID对应的商品信息
			product := new(models.Product)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("Product:%s", orderProductBody.GetProductID()))); err == nil {
				if err := redis.ScanStruct(result, product); err != nil {
					kc.logger.Error("错误: ScanStruct", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("This Product is not exists")
					goto COMPLETE
				}
			}

			//检测商品有效期是否过期， 对彩票竞猜类的商品，有效期内才能下单
			if (product.Expire > 0) && (product.Expire < time.Now().UnixNano()/1e6) {
				kc.logger.Warn("商品有效期过期", zap.Int64("Expire", product.Expire))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Product is expire")
				goto COMPLETE
			}

			//TODO, 余额是否足够扣除

			//将订单转发到商户
			if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", orderProductBody.GetBusinessUser()))); err != nil {
				kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "INCR Error"
				goto COMPLETE
			}

			eRsp := &Msg.RecvMsgEventRsp{
				Scene:        Msg.MessageScene_MsgScene_S2C, //系统消息
				Type:         Msg.MessageType_MsgType_Order, //类型-订单消息
				Body:         req.GetBody(),
				From:         username,                           //谁发的
				FromDeviceId: deviceID,                           //哪个设备发的
				ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
				Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
				Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
				Time:         uint64(time.Now().UnixNano() / 1e6),
			}

			go kc.BroadcastMsgToAllDevices(eRsp, orderProductBody.GetBusinessUser())
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//构造回包消息数据
		if curSeq, err := redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", username))); err != nil {
			kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))

		} else {
			rsp = &Msg.SendMsgRsp{
				Uuid:        req.GetUuid(),
				ServerMsgId: msg.GetID(),
				Seq:         curSeq,
				Time:        uint64(time.Now().UnixNano() / 1e6), //毫秒
			}
			data, _ := proto.Marshal(rsp)
			msg.FillBody(data)
		}
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("HandleOrderMsg: Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("HandleOrderMsg: Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
9-5 对订单进行状态更改
1. 双方都可以更改订单的状态, 只有商户才可以撤单及设置订单完成，用户可以申请撤单及确认收货
*/
func (kc *KafkaClient) HandleChangeOrderState(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64
	var toUser string

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleChangeOrderState start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("ChangeOrderState",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Order.ChangeOrderStateReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		kc.logger.Debug("ChangeOrderState payload",
			zap.String("OrderID", req.OrderBody.GetOrderID()),
			zap.String("ProductID", req.OrderBody.GetProductID()),
			zap.String("BuyUser", req.OrderBody.GetBuyUser()),
			zap.String("OpkBuyUser", req.OrderBody.GetOpkBuyUser()),
			zap.String("BusinessUser", req.OrderBody.GetBusinessUser()),
			zap.String("OpkBusinessUser", req.OrderBody.GetOpkBusinessUser()),
			zap.Int("State", int(req.GetState())),
			zap.Uint64("TimeAt", req.TimeAt),
		)
		if req.OrderBody.GetOrderID() == "" {
			kc.logger.Error("OrderID is empty")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("OrderID is empty")
			goto COMPLETE
		}

		if req.OrderBody.GetProductID() == "" {
			kc.logger.Error("ProductID is empty")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ProductID is empty")
			goto COMPLETE
		}

		if req.OrderBody.GetBuyUser() == "" {
			kc.logger.Error("BuyUser is empty")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("BuyUser is empty")
			goto COMPLETE
		}

		if req.OrderBody.GetOpkBuyUser() == "" {
			kc.logger.Error("OpkBuyUser is empty")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("OpkBuyUser is empty")
			goto COMPLETE
		}

		if req.OrderBody.GetBusinessUser() == "" {
			kc.logger.Error("BusinessUse is empty")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("BusinessUse is empty")
			goto COMPLETE
		}

		if req.OrderBody.GetOpkBusinessUser() == "" {
			kc.logger.Error("OpkBusinessUser is empty")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("OpkBusinessUser is empty")
			goto COMPLETE
		}

		//判断发起方是谁
		if username == req.OrderBody.GetBuyUser() {
			toUser = req.OrderBody.GetBusinessUser()

		} else {
			toUser = req.OrderBody.GetBuyUser()
		}

		//从redis里获取买家信息
		buyerData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", req.OrderBody.GetBuyUser())
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, buyerData); err != nil {

				kc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", req.OrderBody.GetBuyUser())
				goto COMPLETE

			}
		}

		//从redis里获取商户的信息
		businessUserData := new(models.User)
		userKey = fmt.Sprintf("userData:%s", req.OrderBody.GetBusinessUser())
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, businessUserData); err != nil {

				kc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", req.OrderBody.GetBusinessUser())
				goto COMPLETE

			}
		}

		//判断商户是否被封号
		if businessUserData.State == 2 {
			kc.logger.Warn("此商户已被封号", zap.String("businessUser", req.OrderBody.GetBusinessUser()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is blocked[Username=%s]", req.OrderBody.GetBusinessUser())
			goto COMPLETE
		}
		//判断此订单是否已经在商户的有序集合里orders:{账号id}
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("orders:%s", req.OrderBody.GetBusinessUser()), req.OrderBody.GetOrderID()); err == nil {
			if reply == nil {
				//商户不是用户关注
				kc.logger.Error("此订单id不属于此商户",
					zap.String("OrderID", req.OrderBody.GetOrderID()),
				)
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("This orderid is not be geared to BusinessUser:[%s]", req.OrderBody.GetBusinessUser())
				goto COMPLETE
			}

		}
		// 获取ProductID对应的商品信息
		product := new(models.Product)
		if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("Product:%s", req.OrderBody.GetProductID()))); err == nil {
			if err := redis.ScanStruct(result, product); err != nil {
				kc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("This Product is not exists")
				goto COMPLETE
			}
		}

		//获取当前订单的状态
		curState, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("orderInfo:%s", req.OrderBody.GetOrderID()), "orderState"))
		switch Global.OrderState(curState) {
		case Global.OrderState_OS_Undefined:
			_, err = redisConn.Do("HSET",
				fmt.Sprintf("orderInfo:%s", req.OrderBody.GetOrderID()),
				"orderState", Global.OrderState_OS_Prepare, //预审核
			)

		case Global.OrderState_OS_Prepare, //预审核
			Global.OrderState_OS_SendOK,      //订单发送成功
			Global.OrderState_OS_RecvOK,      //订单送达成功
			Global.OrderState_OS_Taked,       //接单成功
			Global.OrderState_OS_Processing,  //订单处理中
			Global.OrderState_OS_Done,        //完成订单
			Global.OrderState_OS_ApplyCancel: //用户申请撤单

			if req.OrderBody.GetBusinessUser() == username { //只有商户才能有权更改订单状态为完成或撤单
				if req.GetState() == Global.OrderState_OS_Done || req.GetState() == Global.OrderState_OS_Cancel {
					//pass
				} else {
					kc.logger.Warn("警告: 只有商户才能有权更改订单状态为撤单或完成")
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("You have not right to change order state")
					goto COMPLETE
				}
			}

			_, err = redisConn.Do("HSET",
				fmt.Sprintf("orderInfo:%s", req.OrderBody.GetOrderID()),
				"orderState", int(req.GetState()), //订单状态
			)

		case Global.OrderState_OS_Confirm: //确认收货
			kc.logger.Warn("警告: 此订单已经确认收货,不能再更改其状态")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("This order is confirmed")
			goto COMPLETE

		case Global.OrderState_OS_Cancel: //撤单
			kc.logger.Warn("警告: 此订单已撤单,不能再更改其状态")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("This order is canceled")
			goto COMPLETE

		}

		//TODO 如果是完成或撤单，需要向钱包发送结算

		//将最新订单状态转发到目标用户
		if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", toUser))); err != nil {
			kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = "INCR Error"
			goto COMPLETE
		}

		orderBodyData, _ := proto.Marshal(req.OrderBody)

		eRsp := &Msg.RecvMsgEventRsp{
			Scene:        Msg.MessageScene_MsgScene_S2C,      //系统消息
			Type:         Msg.MessageType_MsgType_Order,      //类型-订单消息
			Body:         orderBodyData,                      //发起方的body负载
			From:         username,                           //谁发的
			FromDeviceId: deviceID,                           //哪个设备发的
			ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
			Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
			Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
			Time:         uint64(time.Now().UnixNano() / 1e6),
		}

		go kc.BroadcastMsgToAllDevices(eRsp, toUser)
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
		kc.logger.Info("HandleChangeOrderState: Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("HandleChangeOrderState: Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
9-8 商户获取OPK存量
*/
func (kc *KafkaClient) HandleGetPreKeysCount(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var count int

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleGetPreKeysCount start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("GetPreKeysCount",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//从redis里获取当前用户信息
	userData := new(models.User)
	userKey := fmt.Sprintf("userData:%s", username)
	if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
		if err := redis.ScanStruct(result, userData); err != nil {

			kc.logger.Error("错误: ScanStruct", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
			goto COMPLETE

		}
	}

	if userData.UserType != 2 {
		kc.logger.Error("只有商户才能查询OPK存量")
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("UserType is not business type")
		goto COMPLETE
	}

	if count, err = redis.Int(redisConn.Do("ZCOUNT", fmt.Sprintf("prekeys:%s", username), "-inf", "+inf")); err != nil {
		kc.logger.Error("ZCOUNT Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Prekeys is not exists[username=%s]", username)
		goto COMPLETE
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		rsp := &Order.GetPreKeysCountRsp{
			Count: int32(count),
		}
		data, _ := proto.Marshal(rsp)
		msg.FillBody(data)

	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("HandleGetPreKeysCount: Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("HandleGetPreKeysCount: Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}
