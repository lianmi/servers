package nsqBackend

import (
	"encoding/json"
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
7-1 查询某个商户的所有商品信息
1. 根据timeAt增量返回商品信息，首次timeAt请初始化为0，服务器返回全量商品信息，后续采取增量方式更新
2. 如果soldoutProducts不为空，终端根据soldoutProducts移除商品缓存数据
3. 获取商品信息的流程:  发起获取商品信息请求 → 更新本地数据库 → 返回数据给UI

*/
func (nc *NsqClient) HandleQueryProducts(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	rsp := &Order.QueryProductsRsp{
		Products:        make([]*Order.Product, 0),
		SoldoutProducts: make([]string, 0),
		TimeAt:          uint64(time.Now().UnixNano() / 1e6),
	}

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleQueryProducts start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("QueryProducts",
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
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		nc.logger.Debug("QueryProducts  payload",
			zap.String("UserName", req.GetUserName()),
			zap.Uint64("TimeAt", req.GetTimeAt()),
		)
		//从redis里获取当前用户信息
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				nc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
				goto COMPLETE

			}
		}

		//判断此商户是不是用户关注的，如果不是则返回
		/*
			if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Watching:%s", username), req.GetUserName()); err == nil {
				if reply == nil {
					//商户不是用户关注
					nc.logger.Debug("此商户不是用户关注",
						zap.String("UserName", req.GetUserName()),
					)
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("User is not watching[Watching Username=%s]", req.GetUserName())
					goto COMPLETE
				}

			}
		*/

		//获取商户的商品有序集合
		//从redis的有序集合查询出商户的商品信息在时间戳req.GetTimeAt()之后的更新
		productIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Products:%s", req.GetUserName()), req.GetTimeAt(), "+inf"))
		for _, productID := range productIDs {
			product := new(models.Product)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("Product:%s", productID))); err == nil {
				if err := redis.ScanStruct(result, product); err != nil {
					nc.logger.Error("错误: ScanStruct", zap.Error(err))
					continue
				}
			}

			rsp.Products = append(rsp.Products, &Order.Product{
				ProductId:         productID,
				Expire:            uint64(product.Expire),
				ProductName:       product.ProductName,
				ProductType:       Global.ProductType(product.ProductType),
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
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

//7-2 商品上架
func (nc *NsqClient) HandleAddProduct(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	rsp := &Order.AddProductRsp{}

	// var newSeq uint64

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleAddProduct start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("AddProduct",
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
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		nc.logger.Debug("AddProduct  payload",
			zap.String("ProductId", req.GetProduct().ProductId),
			zap.Int("OrderType", int(req.GetOrderType())),
			zap.String("OpkBusinessUser", req.GetOpkBusinessUser()),
			zap.Uint64("Expire", req.GetExpire()),
		)

		if req.GetProduct().ProductId != "" {
			nc.logger.Warn("新的上架商品id必须是空的 ")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ProductId is not empty[Username=%s]", username)
			goto COMPLETE
		}

		if req.GetOrderType() == Global.OrderType_ORT_Normal ||
			req.GetOrderType() == Global.OrderType_ORT_Grabbing ||
			req.GetOrderType() == Global.OrderType_ORT_Walking {
			//符合要求 pass
		} else {
			nc.logger.Warn("新的上架商品所属类型不正确")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("OrderType is not right[Username=%s]", username)
			goto COMPLETE
		}

		//校验过期时间
		if req.GetExpire() > 0 {
			//是否小于当前时间戳
			if int64(req.GetExpire()) < time.Now().UnixNano()/1e6 {
				nc.logger.Warn("Expire小于当前时间戳")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Expire is less than current microsecond[Expire=%d]", req.GetExpire())
				goto COMPLETE
			}

		}

		//生成随机的商品id
		req.Product.ProductId = uuid.NewV4().String()
		rsp.ProductID = req.Product.ProductId
		nc.logger.Debug("新的上架商品ID", zap.String("ProductID", rsp.ProductID))

		//从redis里获取当前用户信息
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				nc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
				goto COMPLETE

			}
		}

		if userData.UserType != int(User.UserType_Ut_Business) {
			nc.logger.Warn("用户不是商户类型，不能上架商品")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is not business type[Username=%s]", username)
			goto COMPLETE
		}

		//上架
		if _, err = redisConn.Do("ZADD", fmt.Sprintf("Products:%s", username), time.Now().UnixNano()/1e6, req.Product.ProductId); err != nil {
			nc.logger.Error("ZADD Error", zap.Error(err))
		}

		product := &models.Product{
			Username:          username,
			ProductID:         req.Product.ProductId,
			Expire:            int64(req.Product.Expire),
			ProductName:       req.Product.ProductName,
			ProductType:       int(req.Product.ProductType),
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
			nc.logger.Error("错误: HMSET ProductInfo", zap.Error(err))
		}

		//保存到MySQL
		if err = nc.SaveProduct(product); err != nil {
			nc.logger.Error("错误: 保存到MySQL失败", zap.Error(err))
		}

		//推送通知给关注的用户
		watchingUsers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("BeWatching:%s", username), "-inf", "+inf"))
		for _, watchingUser := range watchingUsers {

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

			//向所有关注了此商户的用户推送 7-5 新商品上架事件
			go nc.BroadcastSpecialMsgToAllDevices(productData, uint32(Global.BusinessType_Product), uint32(Global.ProductSubType_AddProductEvent), watchingUser)
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
		nc.logger.Debug("7-2 回包")
		data, _ := proto.Marshal(rsp)
		msg.FillBody(data)

	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("7-2 回包 Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("7-2 回包 Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

//7-3 商品编辑更新
func (nc *NsqClient) HandleUpdateProduct(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	// var newSeq uint64

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleUpdateProduct start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("UpdateProduct",
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
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		nc.logger.Debug("UpdateProduct  payload",
			zap.String("ProductId", req.GetProduct().ProductId),
			zap.Int("OrderType", int(req.GetOrderType())),
			// zap.Uint64("Expire", req.GetExpire()),
		)

		if req.GetProduct().ProductId == "" {
			nc.logger.Warn("上架商品id必须非空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ProductId is empty[Username=%s]", username)
			goto COMPLETE
		}

		//从redis里获取当前用户信息
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				nc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
				goto COMPLETE

			}
		}

		if userData.UserType != int(User.UserType_Ut_Business) {
			nc.logger.Warn("用户不是商户类型，不能上架商品")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is not business type[Username=%s]", username)
			goto COMPLETE
		}

		//判断是否是上架
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Products:%s", username), req.Product.ProductId); err == nil {
			if reply == nil {
				//此商品没有上架过
				nc.logger.Warn("此商品没有上架过")
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
			ProductType:       int(req.Product.ProductType),
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
			nc.logger.Error("错误: HMSET Product Info", zap.Error(err))
		}

		//保存到MySQL
		if err = nc.SaveProduct(product); err != nil {
			nc.logger.Error("错误: 保存到MySQL失败", zap.Error(err))
		}

		//推送通知给关注的用户
		watchingUsers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("BeWatching:%s", username), "-inf", "+inf"))
		for _, watchingUser := range watchingUsers {

			//7-6 已有商品的编辑更新事件
			updateProductEventRsp := &Order.UpdateProductEventRsp{
				Username:  username,
				Product:   req.Product,
				OrderType: req.OrderType,
				Expire:    req.Expire,
				TimeAt:    uint64(time.Now().UnixNano() / 1e6),
			}
			productData, _ := proto.Marshal(updateProductEventRsp)

			//向所有关注了此商户的用户推送  7-6 已有商品的编辑更新事件
			go nc.BroadcastSpecialMsgToAllDevices(productData, uint32(Global.BusinessType_Product), uint32(Global.ProductSubType_UpdateProductEvent), watchingUser)
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		msg.FillBody(nil)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

//7-4 商品下架
func (nc *NsqClient) HandleSoldoutProduct(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleSoldoutProduct start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("SoldoutProduct",
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
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		nc.logger.Debug("SoldoutProduct  payload",
			zap.String("ProductId", req.ProductID),
		)

		if req.ProductID == "" {
			nc.logger.Warn("下架商品id必须非空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ProductId is empty[Username=%s]", username)
			goto COMPLETE
		}

		//从redis里获取当前用户信息
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				nc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
				goto COMPLETE

			}
		}

		if userData.UserType != int(User.UserType_Ut_Business) {
			nc.logger.Warn("用户不是商户类型，不能下架商品")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is not business type[Username=%s]", username)
			goto COMPLETE
		}

		//判断是否是上架
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Products:%s", username), req.ProductID); err == nil {
			if reply == nil {
				//此商品没有上架过
				nc.logger.Warn("此商品没有上架过")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Product is not onsell [Username=%s]", username)
				goto COMPLETE
			}

		}
		_, err = redisConn.Do("ZREM", fmt.Sprintf("Products:%s", username), req.ProductID)
		_, err = redisConn.Do("ZADD", fmt.Sprintf("RemoveProducts:%s", username), time.Now().UnixNano()/1e6, req.ProductID)

		//TODO 判断是否存在着此商品id的订单

		//得到此商品的详细信息，如图片等，从阿里云OSS里删除这些文件
		product := new(models.Product)
		if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("Product:%s", req.ProductID))); err == nil {
			if err := redis.ScanStruct(result, product); err != nil {
				nc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "HGETALL Error"
				goto COMPLETE
			}
		}
		if err = nc.DeleteAliyunOssFile(product); err != nil {
			nc.logger.Error("DeleteAliyunOssFile", zap.Error(err))
		}

		//从MySQL删除此商品
		if err = nc.DeleteProduct(req.ProductID, username); err != nil {
			nc.logger.Error("错误: 从MySQL删除对应的req.ProductID失败", zap.Error(err))
		}

		//推送通知给关注的用户
		watchingUsers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("BeWatching:%s", username), "-inf", "+inf"))
		for _, watchingUser := range watchingUsers {
			//7-7 商品下架事件
			soldoutProductEventRsp := &Order.SoldoutProductEventRsp{
				ProductID: req.ProductID,
			}
			productData, _ := proto.Marshal(soldoutProductEventRsp)

			//向所有关注了此商户的用户推送 7-7 商品下架事件
			go nc.BroadcastSpecialMsgToAllDevices(productData, uint32(Global.BusinessType_Product), uint32(Global.ProductSubType_SoldoutProductEvent), watchingUser)
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
		msg.FillBody(nil)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info(" Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

//9-1 商户上传订单DH加密公钥
func (nc *NsqClient) HandleRegisterPreKeys(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleRegisterPreKeys start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("RegisterPreKeys",
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
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		nc.logger.Debug("RegisterPreKeys  payload",
			zap.Strings("PreKeys", req.GetPreKeys()),
		)

		if len(req.GetPreKeys()) == 0 {
			nc.logger.Warn("一次性公钥的数组长度必须大于0")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("PreKeys is empty[Username=%s]", username)
			goto COMPLETE
		}

		//从redis里获取当前用户信息
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				nc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
				goto COMPLETE

			}
		}

		if userData.UserType != int(User.UserType_Ut_Business) {
			nc.logger.Warn("用户不是商户类型，不能上传OPK")
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
				nc.logger.Error("ZADD Error", zap.Error(err))
			}
			nc.logger.Debug("ZADD "+fmt.Sprintf("prekeys:%s", username), zap.String("opk", opk))
		}

		if err = nc.SavePreKeys(prekeys); err != nil {
			nc.logger.Error("SavePreKeys错误", zap.Error(err))
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
		msg.FillBody(nil)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info(" Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

//9-2 获取网点OPK公钥及订单ID
func (nc *NsqClient) HandleGetPreKeyOrderID(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	rsp := &Order.GetPreKeyOrderIDRsp{}
	var count int //OPK有序集合的数量

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleGetPreKeyOrderID start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("GetPreKeyOrderID",
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
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		nc.logger.Debug("GetPreKeyOrderID  payload",
			zap.String("UserName", req.GetUserName()),     //商户
			zap.Int("OrderType", int(req.GetOrderType())), //订单类型
			zap.String("ProducctID", req.GetProductID()),  //商品id
		)

		if req.GetUserName() == "" {
			nc.logger.Warn("商户用户账号不能为空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("UserName is empty[Username=%s]", req.GetUserName())
			goto COMPLETE
		}
		if req.GetProductID() == "" {
			nc.logger.Warn("商品id不能为空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ProductID is empty[ProductID=%s]", req.GetProductID())
			goto COMPLETE
		}

		//从redis里获取当前用户信息
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				nc.logger.Error("错误: ScanStruct", zap.Error(err))
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

				nc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", req.GetUserName())
				goto COMPLETE

			}
		}

		//判断商户是否被封号
		if businessUserData.State == 2 {
			nc.logger.Warn("此商户已被封号", zap.String("businessUser", req.GetUserName()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is blocked[Username=%s]", req.GetUserName())
			goto COMPLETE
		}

		if businessUserData.UserType != int(User.UserType_Ut_Business) {
			nc.logger.Warn("目标用户不是商户类型")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is not business type[Username=%s]", req.GetUserName())
			goto COMPLETE
		}

		// 获取ProductID对应的商品信息
		product := new(models.Product)
		if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("Product:%s", req.GetProductID()))); err == nil {
			if err := redis.ScanStruct(result, product); err != nil {
				nc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("This Product is not exists")
				goto COMPLETE
			}
		}

		//检测商品有效期是否过期， 对彩票竞猜类的商品，有效期内才能下单
		if (product.Expire > 0) && (product.Expire < time.Now().UnixNano()/1e6) {
			nc.logger.Warn("商品有效期过期", zap.Int64("Expire", product.Expire))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Product is expire")
			goto COMPLETE
		}

		// 生成订单ID
		orderID := uuid.NewV4().String()

		opk := ""

		//从商户的prekeys有序集合取出一个opk
		prekeySlice, _ := redis.Strings(redisConn.Do("ZRANGE", fmt.Sprintf("prekeys:%s", req.GetUserName()), 0, 0))
		if len(prekeySlice) > 0 {
			opk = prekeySlice[0]

			//取出后就删除此OPK
			if _, err = redisConn.Do("ZREM", fmt.Sprintf("prekeys:%s", req.GetUserName()), opk); err != nil {
				nc.logger.Error("ZREM Error", zap.Error(err))
			}

		} else {
			nc.logger.Warn("商户的prekeys有序集合无法取出")
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
			nc.logger.Error("ZADD Error", zap.Error(err))
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

		//商户的prekeys有序集合是否少于10个，如果少于，则推送报警，让SDK上传OPK
		if count, err = redis.Int(redisConn.Do("ZCOUNT", fmt.Sprintf("prekeys:%s", req.GetUserName()), "-inf", "+inf")); err != nil {
			nc.logger.Error("ZCOUNT Error", zap.Error(err))
		} else {

			if count < 10 {
				nc.logger.Warn("商户的prekeys存量不足", zap.Int("count", count))

				//查询出商户主设备
				deviceListKey := fmt.Sprintf("devices:%s", req.GetUserName())
				deviceIDSlice, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
				for index, eDeviceID := range deviceIDSlice {
					if index == 0 {
						nc.logger.Debug("查询出商户主设备", zap.Int("index", index), zap.String("eDeviceID", eDeviceID))
						deviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
						jwtToken, _ := redis.String(redisConn.Do("GET", deviceKey))
						nc.logger.Debug("Redis GET ", zap.String("deviceKey", deviceKey), zap.String("jwtToken", jwtToken))

						//向商户主设备推送9-10OPK存量不足事件
						opkAlertMsg := &models.Message{}
						now := time.Now().UnixNano() / 1e6 //毫秒
						opkAlertMsg.UpdateID()

						//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
						opkAlertMsg.BuildRouter("Order", "", "Order.Frontend")
						opkAlertMsg.SetJwtToken(jwtToken)
						opkAlertMsg.SetUserName(req.GetUserName())
						opkAlertMsg.SetDeviceID(string(eDeviceID))
						// opkAlertMsg.SetTaskID(uint32(taskId))
						opkAlertMsg.SetBusinessTypeName("Order")
						opkAlertMsg.SetBusinessType(uint32(Global.BusinessType_Order))            //订单模块
						opkAlertMsg.SetBusinessSubType(uint32(Global.OrderSubType_OPKLimitAlert)) //9-10. 商户OPK存量不足事件

						opkAlertMsg.BuildHeader("OrderService", now)

						//构造负载数据
						resp := &Order.OPKLimitAlertRsp{
							Count: int32(count),
						}
						data, _ := proto.Marshal(resp)
						opkAlertMsg.FillBody(data) //网络包的body，承载真正的业务数据

						opkAlertMsg.SetCode(200) //成功的状态码

						//构建数据完成，向dispatcher发送
						topic := "Order.Frontend"
						rawData, _ := json.Marshal(opkAlertMsg)
						if err := nc.Producer.Public(topic, rawData); err == nil {
							nc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
						} else {
							nc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
						}

						//跳出，不用管从设备
						break

					}
				}

			} else {
				nc.logger.Debug("商户的prekeys存量", zap.Int("count", count))
			}

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
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info(" Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
向目标用户账号的所有端推送消息， 接收端会触发接收消息事件
业务号:  BusinessType_Msg(5)
业务子号:  MsgSubType_RecvMsgEvent(2)
*/
func (nc *NsqClient) BroadcastSystemMsgToAllDevices(rsp *Msg.RecvMsgEventRsp, toUsername string) error {
	data, _ := proto.Marshal(rsp)

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//一次性删除7天前的缓存系统消息
	nTime := time.Now()
	yesTime := nTime.AddDate(0, 0, -7).Unix()

	offLineMsgListKey := fmt.Sprintf("offLineMsgList:%s", toUsername)

	_, err := redisConn.Do("ZREMRANGEBYSCORE", offLineMsgListKey, "-inf", yesTime)

	//Redis里缓存此消息,目的是用户从离线状态恢复到上线状态后同步这些系统消息给用户
	systemMsgAt := time.Now().UnixNano() / 1e6
	if _, err := redisConn.Do("ZADD", offLineMsgListKey, systemMsgAt, rsp.GetServerMsgId()); err != nil {
		nc.logger.Error("ZADD Error", zap.Error(err))
	}

	//系统消息具体内容
	systemMsgKey := fmt.Sprintf("systemMsg:%s:%s", toUsername, rsp.GetServerMsgId())

	_, err = redisConn.Do("HMSET",
		systemMsgKey,
		"Username", toUsername,
		"SystemMsgAt", systemMsgAt,
		"Seq", rsp.Seq,
		"Data", data, //系统消息的数据体
	)

	_, err = redisConn.Do("EXPIRE", systemMsgKey, 7*24*3600) //设置有效期为7天

	//向toUser所有端发送
	deviceListKey := fmt.Sprintf("devices:%s", toUsername)
	deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
	//查询出当前在线所有主从设备
	for _, eDeviceID := range deviceIDSliceNew {

		targetMsg := &models.Message{}
		curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
		curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
		nc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

		targetMsg.UpdateID()
		//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
		targetMsg.BuildRouter("Order", "", "Order.Frontend")

		targetMsg.SetJwtToken(curJwtToken)
		targetMsg.SetUserName(toUsername)
		targetMsg.SetDeviceID(eDeviceID)
		// opkAlertMsg.SetTaskID(uint32(taskId))
		targetMsg.SetBusinessTypeName("Order")
		targetMsg.SetBusinessType(uint32(Global.BusinessType_Msg))           //消息模块
		targetMsg.SetBusinessSubType(uint32(Global.MsgSubType_RecvMsgEvent)) //接收消息事件

		targetMsg.BuildHeader("OrderService", time.Now().UnixNano()/1e6)

		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

		targetMsg.SetCode(200) //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Order.Frontend"
		rawData, _ := json.Marshal(targetMsg)
		if err := nc.Producer.Public(topic, rawData); err == nil {
			nc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
		}

		nc.logger.Info("Broadcast  Msg To All Devices Succeed",
			zap.String("Username:", toUsername),
			zap.String("DeviceID:", curDeviceKey),
			zap.Int64("Now", time.Now().UnixNano()/1e6))

		_ = err

	}

	return nil
}

func (nc *NsqClient) DeleteAliyunOssFile(product *models.Product) error {
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
			nc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic1Small:", product.ProductPic1Small))
		}

	}
	if product.ProductPic1Middle != "" {
		err = bucket.DeleteObject(product.ProductPic1Middle)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic1Middle:", product.ProductPic1Middle))
		}

	}
	if product.ProductPic1Large != "" {
		err = bucket.DeleteObject(product.ProductPic1Large)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic1Large:", product.ProductPic1Large))
		}

	}

	if product.ProductPic2Small != "" {
		err = bucket.DeleteObject(product.ProductPic2Small)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic2Small:", product.ProductPic2Small))
		}

	}
	if product.ProductPic2Middle != "" {
		err = bucket.DeleteObject(product.ProductPic2Middle)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic2Middle:", product.ProductPic2Middle))
		}

	}
	if product.ProductPic2Large != "" {
		err = bucket.DeleteObject(product.ProductPic2Large)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic2Large:", product.ProductPic2Large))
		}

	}

	if product.ProductPic3Small != "" {
		err = bucket.DeleteObject(product.ProductPic3Small)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic3Small:", product.ProductPic3Small))
		}

	}
	if product.ProductPic3Middle != "" {
		err = bucket.DeleteObject(product.ProductPic3Middle)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic3Middle:", product.ProductPic3Middle))
		}

	}
	if product.ProductPic3Large != "" {
		err = bucket.DeleteObject(product.ProductPic3Large)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic3Large:", product.ProductPic3Large))
		}

	}

	if product.Thumbnail != "" {
		err = bucket.DeleteObject(product.Thumbnail)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("Thumbnail:", product.Thumbnail))
		}

	}
	if product.ShortVideo != "" {
		err = bucket.DeleteObject(product.ShortVideo)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("ShortVideo:", product.ShortVideo))
		}

	}

	return nil
}

/*
处理订单消息 5-1
文档是 9-3 下单 处理订单消息，是由ChatService转发过来的
只能是向商户下单
*/
func (nc *NsqClient) HandleOrderMsg(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64

	//经过服务端更改状态后的新的OrderProductBody字节流
	var orderProductBodyData []byte

	rsp := &Msg.SendMsgRsp{}

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleOrderMsg start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("OrderMsg",
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
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		nc.logger.Debug("OrderMsg payload",
			zap.Int32("Scene", int32(req.GetScene())),
			zap.Int32("Type", int32(req.GetType())),
			zap.String("To", req.GetTo()), //商户账户id
			zap.String("Uuid", req.GetUuid()),
			zap.Uint64("SendAt", req.GetSendAt()),
		)

		if req.GetTo() == "" {
			nc.logger.Warn("商户用户账号不能为空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("To is empty[Username=%s]", req.GetTo())
			goto COMPLETE
		}
		if req.GetType() != Msg.MessageType_MsgType_Order {
			nc.logger.Warn("警告，不能处理非订单类型的消息")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Type is not right[Type=%d]", int32(req.GetType()))
			goto COMPLETE
		}

		//从redis里获取当前用户信息
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				nc.logger.Error("错误: ScanStruct", zap.Error(err))
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

				nc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", req.GetTo())
				goto COMPLETE

			}
		}

		//判断商户是否被封号
		if businessUserData.State == 2 {
			nc.logger.Warn("此商户已被封号", zap.String("businessUser", req.GetTo()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is blocked[Username=%s]", req.GetTo())
			goto COMPLETE
		}

		//解包出 OrderProductBody

		var orderProductBody = new(Order.OrderProductBody)
		if err := proto.Unmarshal(req.GetBody(), orderProductBody); err != nil {
			nc.logger.Error("Protobuf Unmarshal OrderProductBody Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Protobuf Unmarshal OrderProductBody Error: %s", err.Error())
			goto COMPLETE

		} else {
			nc.logger.Debug("OrderProductBody payload",
				zap.String("OrderID", orderProductBody.GetOrderID()),
				zap.String("ProductID", orderProductBody.GetProductID()),
				zap.String("BuyUser", orderProductBody.GetBuyUser()),
				zap.String("OpkBuyUser", orderProductBody.GetOpkBuyUser()),
				zap.String("BusinessUser", orderProductBody.GetBusinessUser()),
				zap.String("OpkBusinessUser", orderProductBody.GetOpkBusinessUser()),
				zap.String("Attach", orderProductBody.GetAttach()), //加密的密文
				zap.Int("State", int(orderProductBody.GetState())), //订单状态
			)

			//判断订单id不能为空
			if orderProductBody.GetOrderID() == "" {
				nc.logger.Error("OrderID is empty")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("OrderID is empty")
				goto COMPLETE
			}

			// 判断商品id不能为空
			if orderProductBody.GetProductID() == "" {
				nc.logger.Error("ProductID is empty")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ProductID is empty")
				goto COMPLETE
			}

			//判断买家账号id不能为空
			if orderProductBody.GetBuyUser() == "" {
				nc.logger.Error("BuyUser is empty")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("BuyUser is empty")
				goto COMPLETE
			}

			// 判断买家的OPK不能为空
			if orderProductBody.GetOpkBuyUser() == "" {
				nc.logger.Error("OpkBuyUser is empty")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("OpkBuyUser is empty")
				goto COMPLETE
			}

			//判断商户的账号id不能为空
			if orderProductBody.GetBusinessUser() == "" {
				nc.logger.Error("BusinessUse is empty")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("BusinessUse is empty")
				goto COMPLETE
			}

			// 获取ProductID对应的商品信息
			product := new(models.Product)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("Product:%s", orderProductBody.GetProductID()))); err == nil {
				if err := redis.ScanStruct(result, product); err != nil {
					nc.logger.Error("错误: ScanStruct", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("This Product is not exists")
					goto COMPLETE
				}
			}

			//检测商品有效期是否过期， 对彩票竞猜类的商品，有效期内才能下单
			if (product.Expire > 0) && (product.Expire < time.Now().UnixNano()/1e6) {
				nc.logger.Warn("商品有效期过期", zap.Int64("Expire", product.Expire))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Product is expire")
				goto COMPLETE
			}

			//将订单转发到商户
			if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", orderProductBody.GetBusinessUser()))); err != nil {
				nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "INCR Error"
				goto COMPLETE
			}

			//判断订单状态是不是 OS_Prepare ， 如果是，则改为OS_SendOK
			if Global.OrderState(orderProductBody.GetState()) == Global.OrderState_OS_Prepare {
				orderProductBody.State = Global.OrderState_OS_SendOK
			} else {
				nc.logger.Error("订单状态 Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "OrderProductBody state error, state is not OS_Prepare"
				goto COMPLETE
			}

			orderProductBodyData, _ = proto.Marshal(orderProductBody)

			eRsp := &Msg.RecvMsgEventRsp{
				Scene:        Msg.MessageScene_MsgScene_S2C,      //系统消息
				Type:         Msg.MessageType_MsgType_Order,      //类型-订单消息
				Body:         orderProductBodyData,               //订单载体 OrderProductBody
				From:         username,                           //谁发的
				FromDeviceId: deviceID,                           //哪个设备发的
				Recv:         req.GetTo(),                        //商户账户id
				ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
				Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
				Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
				Time:         uint64(time.Now().UnixNano() / 1e6),
			}
			// 将订单信息缓存在redis里的一个哈希表里, 以 ServerMsgId 对应
			orderProductBodyKey := fmt.Sprintf("OrderProductBody:%s", msg.GetID())
			_, err = redisConn.Do("HMSET",
				orderProductBodyKey,
				"Username", username,
				"OrderID", orderProductBody.GetOrderID(),
				"ProductID", orderProductBody.GetProductID(),
				"BuyUser", orderProductBody.GetBuyUser(),
				"OpkBuyUser", orderProductBody.GetOpkBuyUser(),
				"BusinessUser", orderProductBody.GetBusinessUser(),
				"OpkBusinessUser", orderProductBody.GetOpkBusinessUser(),
				"Attach", orderProductBody.GetAttach(), //加密的密文
				"State", orderProductBody.GetState(), //订单状态
			)

			// 将订单信息缓存在redis里的一个哈希表里, 以 orderID 对应
			orderIDKey := fmt.Sprintf("Order:%s", orderProductBody.GetOrderID())
			_, err = redisConn.Do("HMSET",
				orderIDKey,
				"Username", username, //当前用户
				"OrderID", orderProductBody.GetOrderID(),
				"ProductID", orderProductBody.GetProductID(),
				"BuyUser", orderProductBody.GetBuyUser(), //买家
				"OpkBuyUser", orderProductBody.GetOpkBuyUser(),
				"BusinessUser", orderProductBody.GetBusinessUser(), //商户
				"OpkBusinessUser", orderProductBody.GetOpkBusinessUser(),
				"Attach", orderProductBody.GetAttach(), //加密的密文
				"State", orderProductBody.GetState(), //订单状态
			)

			//向商户发送订单消息
			go nc.BroadcastSystemMsgToAllDevices(eRsp, orderProductBody.GetBusinessUser())
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//构造回包消息数据
		if curSeq, err := redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", username))); err != nil {
			nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
			msg.SetCode(int32(500))                       //状态码
			msg.SetErrorMsg([]byte("INCR userSeq Error")) //错误提示
			msg.FillBody(nil)
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
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("HandleOrderMsg: Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("HandleOrderMsg: Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
9-5 对订单进行状态更改
1. 双方都可以更改订单的状态, 只有商户才可以撤单及设置订单完成，用户可以申请撤单及确认收货
*/
func (nc *NsqClient) HandleChangeOrderState(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64
	var toUsername string

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleChangeOrderState start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("ChangeOrderState",
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
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		nc.logger.Debug("ChangeOrderState payload",
			zap.Int("State", int(req.GetState())),
			zap.Uint64("TimeAt", req.TimeAt),
			zap.String("OrderID", req.OrderBody.GetOrderID()),
			zap.String("ProductID", req.OrderBody.GetProductID()),
			zap.String("BuyUser", req.OrderBody.GetBuyUser()),
			zap.String("OpkBuyUser", req.OrderBody.GetOpkBuyUser()),
			zap.String("BusinessUser", req.OrderBody.GetBusinessUser()),
			zap.String("OpkBusinessUser", req.OrderBody.GetOpkBusinessUser()),
			zap.Float64("OrderTotalAmount", req.OrderBody.GetOrderTotalAmount()),
			zap.String("Attach", req.OrderBody.GetAttach()),
			// zap.String("Userdata", req.OrderBody.GetUserdata()),
		)
		if req.OrderBody.GetOrderID() == "" {
			nc.logger.Error("OrderID is empty")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("OrderID is empty")
			goto COMPLETE
		}

		if req.OrderBody.GetProductID() == "" {
			nc.logger.Error("ProductID is empty")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ProductID is empty")
			goto COMPLETE
		}

		if req.OrderBody.GetBuyUser() == "" {
			nc.logger.Error("BuyUser is empty")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("BuyUser is empty")
			goto COMPLETE
		}

		if req.OrderBody.GetOpkBuyUser() == "" {
			nc.logger.Error("OpkBuyUser is empty")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("OpkBuyUser is empty")
			goto COMPLETE
		}

		if req.OrderBody.GetBusinessUser() == "" {
			nc.logger.Error("BusinessUse is empty")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("BusinessUse is empty")
			goto COMPLETE
		}

		if req.OrderBody.GetOpkBusinessUser() == "" {
			nc.logger.Error("OpkBusinessUser is empty")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("OpkBusinessUser is empty")
			goto COMPLETE
		}

		//判断ToUser方是谁
		if username == req.OrderBody.GetBuyUser() {
			toUsername = req.OrderBody.GetBusinessUser()

		} else {
			toUsername = req.OrderBody.GetBuyUser()
		}

		//从redis里获取买家信息
		buyerData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", req.OrderBody.GetBuyUser())
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, buyerData); err != nil {

				nc.logger.Error("错误: ScanStruct", zap.Error(err))
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

				nc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", req.OrderBody.GetBusinessUser())
				goto COMPLETE

			}
		}

		//判断商户是否被封号
		if businessUserData.State == 2 {
			nc.logger.Warn("此商户已被封号", zap.String("businessUser", req.OrderBody.GetBusinessUser()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("User is blocked[Username=%s]", req.OrderBody.GetBusinessUser())
			goto COMPLETE
		}
		//判断此订单是否已经在商户的有序集合里orders:{账号id}
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("orders:%s", req.OrderBody.GetBusinessUser()), req.OrderBody.GetOrderID()); err == nil {
			if reply == nil {
				//此订单id不属于此商户
				nc.logger.Error("此订单id不属于此商户",
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
				nc.logger.Error("错误: ScanStruct", zap.Error(err))
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
					nc.logger.Warn("警告: 只有商户才能有权更改订单状态为撤单或完成")
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
			nc.logger.Warn("警告: 此订单已经确认收货,不能再更改其状态")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("This order is confirmed")
			goto COMPLETE

		case Global.OrderState_OS_Cancel: //撤单
			nc.logger.Warn("警告: 此订单已撤单,不能再更改其状态")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("This order is canceled")
			goto COMPLETE

		}

		//TODO 如果是完成或撤单，需要向钱包发送结算

		//将最新订单状态转发到目标用户
		if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", toUsername))); err != nil {
			nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
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

		go nc.BroadcastSystemMsgToAllDevices(eRsp, toUsername)
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
		msg.FillBody(nil)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("HandleChangeOrderState: Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("HandleChangeOrderState: Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
9-8 商户获取OPK存量
*/
func (nc *NsqClient) HandleGetPreKeysCount(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var count int

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleGetPreKeysCount start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("GetPreKeysCount",
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

			nc.logger.Error("错误: ScanStruct", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
			goto COMPLETE

		}
	}

	if userData.UserType != 2 {
		nc.logger.Error("只有商户才能查询OPK存量")
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("UserType is not business type")
		goto COMPLETE
	}

	if count, err = redis.Int(redisConn.Do("ZCOUNT", fmt.Sprintf("prekeys:%s", username), "-inf", "+inf")); err != nil {
		nc.logger.Error("ZCOUNT Error", zap.Error(err))
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
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("HandleGetPreKeysCount: Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("HandleGetPreKeysCount: Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
向目标用户账号的所有端推送传入的业务号及子号的消息， 接收端会触发对应事件
传参：
1. data 字节流
2. businessType 业务号
3. businessSubType 业务子号
*/
func (nc *NsqClient) BroadcastSpecialMsgToAllDevices(data []byte, businessType, businessSubType uint32, toUsername string) error {

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//向toUser所有端发送
	deviceListKey := fmt.Sprintf("devices:%s", toUsername)
	deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
	//查询出当前在线所有主从设备
	for _, eDeviceID := range deviceIDSliceNew {

		targetMsg := &models.Message{}
		curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
		curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
		nc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

		targetMsg.UpdateID()
		//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
		targetMsg.BuildRouter("Order", "", "Order.Frontend")

		targetMsg.SetJwtToken(curJwtToken)
		targetMsg.SetUserName(toUsername)
		targetMsg.SetDeviceID(eDeviceID)
		// opkAlertMsg.SetTaskID(uint32(taskId))
		targetMsg.SetBusinessTypeName("Order")
		targetMsg.SetBusinessType(businessType)       //业务号
		targetMsg.SetBusinessSubType(businessSubType) //业务子号

		targetMsg.BuildHeader("OrderService", time.Now().UnixNano()/1e6)

		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

		targetMsg.SetCode(200) //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Order.Frontend"
		rawData, _ := json.Marshal(targetMsg)
		if err := nc.Producer.Public(topic, rawData); err == nil {
			nc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
		}

		nc.logger.Info("Broadcast Msg To All Devices Succeed",
			zap.String("Username:", toUsername),
			zap.String("DeviceID:", curDeviceKey),
			zap.Int64("Now", time.Now().UnixNano()/1e6))

	}

	return nil
}

/*
9-11 确认支付订单, 转发到钱包服务进行下一步处理
*/
func (nc *NsqClient) HandlePayOrder(msg *models.Message) error {
	msg.BuildRouter("Wallet", "", "Wallet.Backend")
	topic := "Wallet.Backend"

	rawData, _ := json.Marshal(msg)
	go nc.Producer.Public(topic, rawData)
	return nil
}
