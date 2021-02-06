package nsqMq

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/api/proto/global"
	Global "github.com/lianmi/servers/api/proto/global"
	Msg "github.com/lianmi/servers/api/proto/msg"
	Order "github.com/lianmi/servers/api/proto/order"
	User "github.com/lianmi/servers/api/proto/user"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	LMCommon "github.com/lianmi/servers/internal/common"
	LMCError "github.com/lianmi/servers/internal/pkg/lmcerror"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/util/crypt"
	"github.com/lianmi/servers/util/mathtool"

	"google.golang.org/protobuf/proto"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	// "github.com/pkg/errors"
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

	rsp := &Order.QueryProductsRsp{
		Products:        make([]*Order.Product, 0),
		SoldoutProducts: make([]string, 0),
		TimeAt:          uint64(time.Now().UnixNano() / 1e6),
	}

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
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
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		nc.logger.Debug("QueryProducts  payload",
			zap.String("UserName", req.UserName),
			zap.Uint64("TimeAt", req.TimeAt),
		)

		//获取商户的商品有序集合
		//从redis的有序集合查询出商户的商品信息在时间戳req.TimeAt之后的更新
		productIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Products:%s", req.UserName), req.TimeAt, "+inf"))
		for _, productID := range productIDs {
			productInfo := new(models.ProductInfo)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("Product:%s", productID))); err == nil {
				if err := redis.ScanStruct(result, productInfo); err != nil {
					nc.logger.Error("错误: ScanStruct", zap.Error(err))
					continue
				}
			}

			var thumbnail string
			if productInfo.ShortVideo != "" {

				thumbnail = LMCommon.OSSUploadPicPrefix + productInfo.ShortVideo + "?x-oss-process=video/snapshot,t_500,f_jpg,w_800,h_600"
			}

			oProduct := &Order.Product{
				ProductId:         productInfo.ProductID,                       //商品ID
				Expire:            uint64(productInfo.Expire),                  //商品过期时间
				ProductName:       productInfo.ProductName,                     //商品名称
				ProductType:       Global.ProductType(productInfo.ProductType), //商品种类类型  枚举
				SubType:           Global.LotteryType_LT_Shuangseqiu,           //TODO  暂时全部都是双色球
				ProductDesc:       productInfo.ProductDesc,                     //商品详细介绍
				ShortVideo:        productInfo.ShortVideo,                      //商品短视频
				Thumbnail:         thumbnail,                                   //商品短视频缩略图
				Price:             productInfo.Price,                           //价格
				LeftCount:         productInfo.LeftCount,                       //库存数量
				Discount:          productInfo.Discount,                        //折扣 实际数字，例如: 0.95, UI显示为九五折
				DiscountDesc:      productInfo.DiscountDesc,                    //折扣说明
				DiscountStartTime: uint64(productInfo.DiscountStartTime),       //折扣开始时间
				DiscountEndTime:   uint64(productInfo.DiscountEndTime),         //折扣结束时间
				AllowCancel:       productInfo.AllowCancel,                     //是否允许撤单， 默认是可以，彩票类的不可以
			}
			if productInfo.ProductPic1Large != "" {
				// 动态拼接
				oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
					Small:  LMCommon.OSSUploadPicPrefix + productInfo.ProductPic1Large + "?x-oss-process=image/resize,w_50/quality,q_50",
					Middle: LMCommon.OSSUploadPicPrefix + productInfo.ProductPic1Large + "?x-oss-process=image/resize,w_100/quality,q_100",
					Large:  LMCommon.OSSUploadPicPrefix + productInfo.ProductPic1Large,
				})
			}

			if productInfo.ProductPic2Large != "" {
				// 动态拼接
				oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
					Small:  LMCommon.OSSUploadPicPrefix + productInfo.ProductPic2Large + "?x-oss-process=image/resize,w_50/quality,q_50",
					Middle: LMCommon.OSSUploadPicPrefix + productInfo.ProductPic2Large + "?x-oss-process=image/resize,w_100/quality,q_100",
					Large:  LMCommon.OSSUploadPicPrefix + productInfo.ProductPic2Large,
				})
			}

			if productInfo.ProductPic3Large != "" {
				// 动态拼接
				oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
					Small:  LMCommon.OSSUploadPicPrefix + productInfo.ProductPic3Large + "?x-oss-process=image/resize,w_50/quality,q_50",
					Middle: LMCommon.OSSUploadPicPrefix + productInfo.ProductPic3Large + "?x-oss-process=image/resize,w_100/quality,q_100",
					Large:  LMCommon.OSSUploadPicPrefix + productInfo.ProductPic3Large,
				})
			}

			if productInfo.DescPic1 != "" {
				oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+productInfo.DescPic1)
			}

			if productInfo.DescPic2 != "" {
				oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+productInfo.DescPic2)
			}

			if productInfo.DescPic3 != "" {
				oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+productInfo.DescPic3)
			}

			if productInfo.DescPic4 != "" {
				oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+productInfo.DescPic4)
			}

			if productInfo.DescPic5 != "" {
				oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+productInfo.DescPic5)
			}

			if productInfo.DescPic6 != "" {
				oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+productInfo.DescPic6)

			}

			rsp.Products = append(rsp.Products, oProduct)
		}

		//获取商户的下架soldoutProducts
		soldoutProductIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("SoldoutProducts:%s", req.UserName), req.TimeAt, "+inf"))
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
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
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
	var data []byte
	errorCode := 200

	var productId string
	var productPic1Small, productPic1Middle, productPic1Large string
	var productPic2Small, productPic2Middle, productPic2Large string
	var productPic3Small, productPic3Middle, productPic3Large string
	var shortVideo, thumbnail string
	var product = &models.Product{}

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
	if err = proto.Unmarshal(body, &req); err != nil {
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		if len(req.Product.ProductPics) >= 1 {
			//大图
			productPic1Large = req.Product.ProductPics[0].Large

		}
		if len(req.Product.ProductPics) >= 2 {
			//大图
			productPic2Large = req.Product.ProductPics[1].Large
		}

		if len(req.Product.ProductPics) >= 3 {
			//大图
			productPic3Large = req.Product.ProductPics[2].Large
		}

		//如果有短视频，则组装缩略图
		if req.Product.ShortVideo != "" {
			shortVideo = req.Product.ShortVideo
		} else {
			shortVideo = ""
		}
		nc.logger.Debug("AddProduct payload",
			zap.String("ProductId", req.Product.ProductId),
			zap.Int("OrderType", int(req.OrderType)),
			zap.String("ProductPic1Large", productPic1Large),
			zap.String("ProductPic2Large", productPic2Large),
			zap.String("ProductPic3Large", productPic3Large),
			zap.String("OpkBusinessUser", req.OpkBusinessUser),
			zap.Uint64("Expire", req.Expire),
		)

		if req.Product.ProductId != "" {
			nc.logger.Warn("新的上架商品id必须是空的")
			errorCode = LMCError.OrderModProductIDNotEmpty //错误码
			goto COMPLETE
		}

		if req.OrderType == Global.OrderType_ORT_Normal ||
			req.OrderType == Global.OrderType_ORT_Grabbing ||
			req.OrderType == Global.OrderType_ORT_Walking {
			//符合要求 pass
		} else {
			nc.logger.Warn("新的上架商品所属类型不正确")
			errorCode = LMCError.OrderModProductTypeError //错误码
			goto COMPLETE
		}

		//校验过期时间
		if req.Expire > 0 {
			//是否小于当前时间戳
			if int64(req.Expire) < time.Now().UnixNano()/1e6 {
				nc.logger.Warn("过期时间小于当前时间戳")
				errorCode = LMCError.OrderModProductTypeError //错误码
				goto COMPLETE
			}

		}

		//生成随机的商品id
		productId = uuid.NewV4().String()
		req.Product.ProductId = productId
		rsp := &Order.AddProductRsp{
			ProductID: productId,
		}
		data, _ = proto.Marshal(rsp)

		nc.logger.Debug("新的上架商品ID", zap.String("ProductID", rsp.ProductID))

		//从redis里获取当前用户信息
		userKey := fmt.Sprintf("userData:%s", username)
		userType, _ := redis.Int(redisConn.Do("HGET", userKey, "UserType"))

		if userType != int(User.UserType_Ut_Business) {
			nc.logger.Warn("用户不是商户类型，不能上架商品")
			errorCode = LMCError.OrderModAddProductUserTypeError //错误码
			goto COMPLETE
		}

		//上架
		if _, err = redisConn.Do("ZADD", fmt.Sprintf("Products:%s", username), time.Now().UnixNano()/1e6, req.Product.ProductId); err != nil {
			nc.logger.Error("ZADD Error", zap.Error(err))
		}

		product = &models.Product{
			ProductInfo: models.ProductInfo{
				ProductID:         productId,
				Username:          username,
				Expire:            int64(req.Product.Expire),
				ProductName:       req.Product.ProductName,
				ProductType:       int(req.Product.ProductType),
				ProductDesc:       req.Product.ProductDesc,
				ProductPic1Large:  productPic1Large,
				ProductPic2Large:  productPic2Large,
				ProductPic3Large:  productPic3Large,
				ShortVideo:        shortVideo,
				Price:             req.Product.Price,
				LeftCount:         req.Product.LeftCount,
				Discount:          req.Product.Discount,
				DiscountDesc:      req.Product.DiscountDesc,
				DiscountStartTime: int64(req.Product.DiscountStartTime),
				DiscountEndTime:   int64(req.Product.DiscountEndTime),
				AllowCancel:       req.Product.AllowCancel,
			},
		}

		if len(req.Product.DescPics) >= 1 {
			product.ProductInfo.DescPic1 = req.Product.DescPics[0]
		}
		if len(req.Product.DescPics) >= 2 {
			product.ProductInfo.DescPic2 = req.Product.DescPics[1]
		}
		if len(req.Product.DescPics) >= 3 {
			product.ProductInfo.DescPic3 = req.Product.DescPics[2]
		}
		if len(req.Product.DescPics) >= 4 {
			product.ProductInfo.DescPic4 = req.Product.DescPics[3]
		}
		if len(req.Product.DescPics) >= 5 {
			product.ProductInfo.DescPic5 = req.Product.DescPics[4]
		}
		if len(req.Product.DescPics) >= 6 {
			product.ProductInfo.DescPic6 = req.Product.DescPics[5]
		}

		nc.logger.Debug("Product字段",
			zap.String("Username", product.ProductInfo.Username),
			zap.String("ProductId", product.ProductInfo.ProductID),
			zap.Int64("Expire", product.ProductInfo.Expire),
			zap.String("ProductName", product.ProductInfo.ProductName),
			zap.Int("ProductType", product.ProductInfo.ProductType),
			zap.String("ProductDesc", product.ProductInfo.ProductDesc),
			zap.String("ProductPic1Large", product.ProductInfo.ProductPic1Large),
			zap.Bool("AllowCancel", product.ProductInfo.AllowCancel),
		)
		//保存到MySQL
		if err = nc.service.AddProduct(product); err != nil {
			nc.logger.Error("错误: 增加到MySQL失败", zap.Error(err))
			errorCode = LMCError.DataBaseError
			goto COMPLETE
		}

		if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("Product:%s", req.Product.ProductId)).AddFlat(product.ProductInfo)...); err != nil {
			nc.logger.Error("错误: HMSET ProductInfo", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {

		nc.logger.Debug("7-2 回包")

		msg.FillBody(data)

	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
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

	go func() {
		//推送通知给关注的用户
		watchingUsers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("BeWatching:%s", username), "-inf", "+inf"))
		for _, watchingUser := range watchingUsers {
			if shortVideo != "" {

				thumbnail = LMCommon.OSSUploadPicPrefix + shortVideo + "?x-oss-process=video/snapshot,t_500,f_jpg,w_800,h_600"
			}
			//7-5 新商品上架事件 将商品信息序化
			oProduct := &Order.Product{
				ProductId:   productId,                                   //商品ID
				Expire:      uint64(req.Product.Expire),                  //商品过期时间
				ProductName: req.Product.ProductName,                     //商品名称
				ProductType: Global.ProductType(req.Product.ProductType), //商品种类类型  枚举
				//TODO  暂时全部都是双色球
				SubType:     Global.LotteryType_LT_Shuangseqiu,
				ProductDesc: req.Product.ProductDesc, //商品详细介绍
				ShortVideo:  shortVideo,
				Thumbnail:   thumbnail,

				Price:             req.Product.Price,                     //价格
				LeftCount:         req.Product.LeftCount,                 //库存数量
				Discount:          req.Product.Discount,                  //折扣 实际数字，例如: 0.95, UI显示为九五折
				DiscountDesc:      req.Product.DiscountDesc,              //折扣说明
				DiscountStartTime: uint64(req.Product.DiscountStartTime), //折扣开始时间
				DiscountEndTime:   uint64(req.Product.DiscountEndTime),   //折扣结束时间
				AllowCancel:       req.Product.AllowCancel,               //是否允许撤单， 默认是可以，彩票类的不可以
			}

			if productPic1Large != "" {

				oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
					Small:  productPic1Small,
					Middle: productPic1Middle,
					Large:  productPic1Large,
				})
			}

			if productPic2Large != "" {

				oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
					Small:  productPic2Small,
					Middle: productPic2Middle,
					Large:  productPic2Large,
				})
			}

			if productPic3Large != "" {
				oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
					Small:  productPic3Small,
					Middle: productPic3Middle,
					Large:  productPic3Large,
				})
			}

			for _, pic := range req.Product.DescPics {

				if pic != "" {
					oProduct.DescPics = append(oProduct.DescPics, pic)
				}
			}

			addProductEventRsp := &Order.AddProductEventRsp{
				Username:    username,            //商户用户账号id
				Product:     oProduct,            //商品数据
				OrderType:   req.OrderType,       //订单类型，必填
				OpkBusiness: req.OpkBusinessUser, //商户的协商公钥，适用于任务类
				Expire:      req.Expire,          //商品过期时间
				TimeAt:      uint64(time.Now().UnixNano() / 1e6),
			}
			productData, _ := proto.Marshal(addProductEventRsp)

			//向所有关注了此商户的用户推送 7-5 新商品上架事件
			nc.BroadcastSpecialMsgToAllDevices(productData, uint32(Global.BusinessType_Product), uint32(Global.ProductSubType_AddProductEvent), watchingUser)
		}
	}()

	return nil
}

//7-3 商品编辑更新
func (nc *NsqClient) HandleUpdateProduct(msg *models.Message) error {
	var err error
	errorCode := 200

	var productPic1Large string
	var productPic2Large string
	var productPic3Large string
	var shortVideo string

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

		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		nc.logger.Debug("UpdateProduct  payload",
			zap.String("ProductId", req.Product.ProductId),
			zap.Int("OrderType", int(req.OrderType)),
			// zap.Uint64("Expire", req.Expire),
		)

		if req.Product.ProductId == "" {
			nc.logger.Warn("上架商品id必须非空")
			errorCode = LMCError.ProductIDIsEmptError
			goto COMPLETE
		}

		//从redis里获取当前用户信息
		userKey := fmt.Sprintf("userData:%s", username)
		userType, _ := redis.Int(redisConn.Do("HGET", userKey, "UserType"))

		if userType != int(User.UserType_Ut_Business) {
			nc.logger.Warn("用户不是商户类型，不能上架商品")
			errorCode = LMCError.OrderModAddProductUserTypeError
			goto COMPLETE
		}

		//判断是否是上架
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Products:%s", username), req.Product.ProductId); err == nil {
			if reply == nil {
				//此商品没有上架过
				nc.logger.Warn("此商品没有上架过")
				errorCode = LMCError.OrderModAddProductNotOnSellError
				goto COMPLETE
			}

		}
		//将3张图片的url组装为真正的url
		if len(req.Product.ProductPics) >= 1 {
			//大图
			productPic1Large = req.Product.ProductPics[0].Large

		}
		if len(req.Product.ProductPics) >= 2 {
			//大图
			productPic2Large = req.Product.ProductPics[1].Large
		}

		if len(req.Product.ProductPics) >= 3 {
			//大图
			productPic3Large = req.Product.ProductPics[2].Large
		}

		//如果有短视频，则组装缩略图
		if req.Product.ShortVideo != "" {
			shortVideo = req.Product.ShortVideo
		} else {
			shortVideo = ""
		}

		product := &models.Product{
			ProductInfo: models.ProductInfo{
				Username:    username,
				ProductID:   req.Product.ProductId,
				Expire:      int64(req.Product.Expire),
				ProductName: req.Product.ProductName,
				ProductType: int(req.Product.ProductType),
				ProductDesc: req.Product.ProductDesc,

				ProductPic1Large: productPic1Large,

				ProductPic2Large: productPic2Large,

				ProductPic3Large: productPic3Large,

				ShortVideo:        shortVideo,
				Price:             req.Product.Price,
				LeftCount:         req.Product.LeftCount,
				Discount:          req.Product.Discount,
				DiscountDesc:      req.Product.DiscountDesc,
				DiscountStartTime: int64(req.Product.DiscountStartTime),
				DiscountEndTime:   int64(req.Product.DiscountEndTime),
				AllowCancel:       req.Product.AllowCancel,
			},
		}

		if len(req.Product.DescPics) >= 1 {
			product.ProductInfo.DescPic1 = req.Product.DescPics[0]
		}
		if len(req.Product.DescPics) >= 2 {
			product.ProductInfo.DescPic2 = req.Product.DescPics[1]
		}
		if len(req.Product.DescPics) >= 3 {
			product.ProductInfo.DescPic3 = req.Product.DescPics[2]
		}
		if len(req.Product.DescPics) >= 4 {
			product.ProductInfo.DescPic4 = req.Product.DescPics[3]
		}
		if len(req.Product.DescPics) >= 5 {
			product.ProductInfo.DescPic5 = req.Product.DescPics[4]
		}
		if len(req.Product.DescPics) >= 6 {
			product.ProductInfo.DescPic6 = req.Product.DescPics[5]
		}

		//保存到MySQL
		if err = nc.service.UpdateProduct(product); err != nil {
			nc.logger.Error("错误: 保存到MySQL失败", zap.Error(err))
			errorCode = LMCError.DataBaseError
			goto COMPLETE
		}

		if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("Product:%s", req.Product.ProductId)).AddFlat(product.ProductInfo)...); err != nil {
			nc.logger.Error("错误: HMSET Product Info", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		}

		go func() {
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
				nc.BroadcastSpecialMsgToAllDevices(productData, uint32(Global.BusinessType_Product), uint32(Global.ProductSubType_UpdateProductEvent), watchingUser)
			}

		}()

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		msg.FillBody(nil)
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
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
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		nc.logger.Debug("SoldoutProduct  payload",
			zap.String("ProductId", req.ProductID),
		)

		if req.ProductID == "" {
			nc.logger.Warn("下架商品id必须非空")
			errorCode = LMCError.ProductIDIsEmptError
			goto COMPLETE
		}

		//从redis里获取当前用户信息
		userKey := fmt.Sprintf("userData:%s", username)
		userType, _ := redis.Int(redisConn.Do("HGET", userKey, "UserType"))

		if userType != int(User.UserType_Ut_Business) {
			nc.logger.Warn("用户不是商户类型，不能下架商品")
			errorCode = LMCError.OrderModAddProductUserTypeError
			goto COMPLETE
		}

		//判断是否是上架
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Products:%s", username), req.ProductID); err == nil {
			if reply == nil {
				//此商品没有上架过
				nc.logger.Warn("此商品没有上架过")
				errorCode = LMCError.OrderModAddProductNotOnSellError
				goto COMPLETE
			}

		}
		_, err = redisConn.Do("ZREM", fmt.Sprintf("Products:%s", username), req.ProductID)
		_, err = redisConn.Do("ZADD", fmt.Sprintf("SoldoutProducts:%s", username), time.Now().UnixNano()/1e6, req.ProductID)

		//TODO 判断是否存在着此商品id的订单

		//得到此商品的详细信息，如图片等，从阿里云OSS里删除这些文件
		productInfo := new(models.ProductInfo)
		if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("Product:%s", req.ProductID))); err == nil {
			if err := redis.ScanStruct(result, productInfo); err != nil {
				nc.logger.Error("错误: ScanStruct", zap.Error(err))
				errorCode = LMCError.RedisError
				goto COMPLETE
			}
		}
		if err = nc.DeleteAliyunOssFile(productInfo); err != nil {
			nc.logger.Error("DeleteAliyunOssFile", zap.Error(err))
		}

		//从MySQL删除此商品
		if err = nc.service.DeleteProduct(req.ProductID, username); err != nil {
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
			nc.BroadcastSpecialMsgToAllDevices(productData, uint32(Global.BusinessType_Product), uint32(Global.ProductSubType_SoldoutProductEvent), watchingUser)
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
		msg.FillBody(nil)
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
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
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		nc.logger.Debug("RegisterPreKeys  payload",
			zap.Strings("PreKeys", req.PreKeys),
		)

		if len(req.PreKeys) == 0 {
			nc.logger.Warn("一次性公钥的数组长度必须大于0")
			errorCode = LMCError.RegisterPreKeysArrayEmptyError
			goto COMPLETE
		}

		//从redis里获取当前用户信息
		userKey := fmt.Sprintf("userData:%s", username)
		userType, _ := redis.Int(redisConn.Do("HGET", userKey, "UserType"))

		if userType != int(User.UserType_Ut_Business) {
			nc.logger.Warn("用户不是商户类型，不能上传OPK")
			errorCode = LMCError.RegisterPreKeysNotBusinessTypeError
			goto COMPLETE
		}

		//opk入库
		prekeys := make([]*models.Prekey, 0)
		for _, opk := range req.PreKeys {
			prekeys = append(prekeys, &models.Prekey{
				Type:      0,
				Username:  username,
				Publickey: opk,
			})

			//保存到redis里prekeys:{username}
			if _, err := redisConn.Do("ZADD", fmt.Sprintf("prekeys:%s", username), time.Now().UnixNano()/1e6, opk); err != nil {
				nc.logger.Error("ZADD Error", zap.Error(err))
			}
			nc.logger.Debug("ZADD "+fmt.Sprintf("prekeys:%s", username), zap.String("opk", opk))
		}

		if err = nc.service.AddPreKeys(prekeys); err != nil {
			nc.logger.Error("AddPreKeys错误", zap.Error(err))
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
		msg.FillBody(nil)
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
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
生成某个订单的服务费订单ID并发送给买家
*/
func (nc *NsqClient) SendChargeOrderIDToBuyer(sdkUuid string, isVip bool, orderProductBody *Order.OrderProductBody) error {
	var err error
	var charge float64
	var newSeq uint64

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//根据用户是否是Vip计算手续费
	charge, err = nc.CalculateCharge(isVip, orderProductBody.OrderTotalAmount)

	//服务费的商品ID
	systemChargeProductID, err := redis.String(redisConn.Do("GET", "SystemChargeProductID"))
	if err != nil {
		nc.logger.Error("SendChargeOrderIDToBuyer GET error")
		return err
	}

	// 生成charge订单ID
	chargeOrderID := uuid.NewV4().String()

	var orignOrder = &models.OrignOrder{
		orderProductBody.OrderID, //真实的商品 订单id
	}
	attachBase := new(models.AttachBase)
	attachBase.BodyType = 100 //约定为服务费的type
	temp, _ := orignOrder.ToJson()
	attachBase.Body = base64.StdEncoding.EncodeToString([]byte(temp)) //约定，购买会员 及 服务费的attach的body部分，都是base64

	attach, _ := attachBase.ToJson()
	attachHex := hex.EncodeToString([]byte(attach))

	// 将服务费数据保存到MySQL
	err = nc.service.SaveChargeHistory(&models.ChargeHistory{
		BuyerUsername:    orderProductBody.BuyUser,          //买家
		BusinessUsername: LMCommon.ChargeBusinessUsername,   //系统商户
		ChargeProductID:  systemChargeProductID,             //服务费的商品D
		ChargeOrderID:    chargeOrderID,                     //本次服务费的订单ID
		BusinessOrderID:  orderProductBody.OrderID,          //商品订单ID, 买家支付的订单ID
		OrderTotalAmount: orderProductBody.OrderTotalAmount, //人民币格式的订单总金额
		IsVip:            isVip,                             //是否是Vip用户
		Rate:             LMCommon.Rate,                     //费率
		ChargeAmount:     mathtool.FloatRound(charge, 2),    //服务费, 取小数点后两位的精度
		IsPayed:          false,
	})

	// 将服务费订单ID信息缓存在redis里的一个哈希表里(Order:{订单ID}), 以 orderID 对应

	//上链服务费的附件类型
	attachType := int(Msg.AttachType_AttachType_BlockServiceCharge)
	_, err = redisConn.Do("HMSET",
		fmt.Sprintf("Order:%s", chargeOrderID),        //charge订单id
		"OrderType", int(Global.OrderType_ORT_Server), //订单类型是服务费
		"ProductID", systemChargeProductID, //服务费的商品ID
		"BuyUser", orderProductBody.BuyUser, //买家
		"OpkBuyUser", "",
		"BusinessUser", LMCommon.ChargeBusinessUsername, //系统商户
		"OpkBusinessUser", "",
		"OrderTotalAmount", mathtool.FloatRound(charge, 2), //服务费, 取小数点后两位的精度
		"AttachType", attachType, //附件类型
		"Attach", attachHex, //hex
		"State", int(Global.OrderState_OS_Prepare), //订单状态
		"IsPayed", LMCommon.REDISFALSE, //此charge订单支付状态， true- 支付完成，false-未支付
		"CreateAt", uint64(time.Now().UnixNano()/1e6), //毫秒
	)
	if err != nil {
		nc.logger.Error("SendChargeOrderIDToBuyer HMSET error")
		return err
	}

	//TODO 将服务费订单ID 发给买家
	chargeOrderProductBody := &Order.OrderProductBody{
		OrderID:          chargeOrderID,                   //charge订单id
		OrderType:        global.OrderType_ORT_Server,     //服务端发起的收费
		ProductID:        systemChargeProductID,           //服务费的商品ID
		BuyUser:          orderProductBody.BuyUser,        //发起订单的用户id
		OpkBuyUser:       "",                              //买家的协商公钥 留空
		BusinessUser:     LMCommon.ChargeBusinessUsername, //商户的用户id, 暂定为id10
		OpkBusinessUser:  "",                              //商户的协商公钥 留空
		OrderTotalAmount: mathtool.FloatRound(charge, 2),  //服务费, 取小数点后两位的精度
		AttachType:       Msg.AttachType_AttachType_BlockServiceCharge,
		Attach:           attachHex,                   // hex json格式的内容 , 由 ui 层处理 sdk 仅透传  传输会进过sdk处理,  这里存放的是真正的订单ID
		State:            Global.OrderState_OS_RecvOK, //订单的状态
	}
	nc.logger.Debug("chargeOrderProductBody", zap.String("chargeOrderID", chargeOrderID), zap.Float64("OrderTotalAmount", mathtool.FloatRound(charge, 2)))

	if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", LMCommon.ChargeBusinessUsername))); err != nil {
		nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
		return err
	}
	chargeOrderProductBodyData, _ := proto.Marshal(chargeOrderProductBody)

	//新建服务费订单消息
	chargeMsg := &models.Message{}

	chargeMsg.UpdateID()
	chargeMsg.SetTaskID(0)

	eRsp := &Msg.RecvMsgEventRsp{
		Scene:        Msg.MessageScene_MsgScene_S2C,   //系统消息
		Type:         Msg.MessageType_MsgType_Order,   //类型- 手续费 消息
		Body:         chargeOrderProductBodyData,      //订单载体
		From:         LMCommon.ChargeBusinessUsername, //谁发的, 暂定为 id10
		FromDeviceId: "",                              //哪个设备发的
		Recv:         orderProductBody.BuyUser,        //商户账户id, 暂定为 id3
		ServerMsgId:  chargeMsg.GetID(),               //服务器分配的消息ID
		Seq:          newSeq,                          //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
		Uuid:         sdkUuid,                         //客户端分配的消息ID，SDK生成的消息id，用来标识UI层的消息
		Time:         uint64(time.Now().UnixNano() / 1e6),
	}

	// 将订单信息OrderProductBody数据缓存在redis里的一个哈希表里, 以 ServerMsgId 对应
	orderProductBodyKey := fmt.Sprintf("OrderProductBody:%s", chargeMsg.GetID())
	_, err = redisConn.Do("HMSET",
		orderProductBodyKey,
		"Username", LMCommon.ChargeBusinessUsername, //暂定为 id10
		"OrderID", chargeOrderID,
		"ProductID", systemChargeProductID,
		"BuyUser", orderProductBody.BuyUser,
		"OpkBuyUser", "", //留空
		"BusinessUser", LMCommon.ChargeBusinessUsername, //商户收款暂定为id10
		"OpkBusinessUser", "", //留空
		"OrderTotalAmount", mathtool.FloatRound(charge, 2), //服务费, 取小数点后两位的精度
		"Attach", attachHex, //hex 真正订单ID，UI负责解析并合并支付
		"State", Global.OrderState_OS_RecvOK, //订单的状态
	)

	//向买家发送 服务费 订单ID消息
	go func() {
		// time.Sleep(100 * time.Millisecond)
		nc.logger.Debug("延时100ms向买家发送 服务费 订单ID消息, 5-2",
			zap.String("to", orderProductBody.BuyUser),
			zap.Int("State", int(orderProductBody.State)),
		)
		nc.BroadcastOrderMsgToAllDevices(eRsp, orderProductBody.BuyUser)
	}()

	return nil

}

//9-2 获取网点OPK公钥及订单ID
func (nc *NsqClient) HandleGetPreKeyOrderID(msg *models.Message) error {
	var err error
	// var sdkUUID string //SDK 传上来的 UUID
	errorCode := 200

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
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		nc.logger.Debug("GetPreKeyOrderID payload",
			zap.String("UserName", req.UserName),     //商户, 当userName=id3时，表示购买VIP会员
			zap.Int("OrderType", int(req.OrderType)), //订单类型
			zap.String("ProducctID", req.ProductID),  //商品id
		)

		if req.ProductID == "" {
			nc.logger.Warn("商品id不能为空")
			errorCode = LMCError.GetPreKeyOrderIDEmptyProductIDError
			goto COMPLETE
		}

		if req.UserName == "" {
			nc.logger.Warn("商户用户账号不能为空")
			errorCode = LMCError.BusinessUsernameIsEmptyError
			goto COMPLETE
		}

		//从redis里获取目标商户的信息
		businessUserState, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", req.UserName), "State"))
		businessUserType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", req.UserName), "UserType"))

		//判断商户是否被封号
		if businessUserState == 2 {
			nc.logger.Warn("此商户已被封号", zap.String("businessUser", req.UserName))
			errorCode = LMCError.BusinessUserIsBlockedError
			goto COMPLETE
		}

		if businessUserType != int(User.UserType_Ut_Business) {
			nc.logger.Warn("目标用户不是商户类型")
			errorCode = LMCError.TargetUserIsNotBusinessTypeError
			goto COMPLETE
		}

		// 获取ProductID对应的商品信息里的过期时间
		expire, _ := redis.Int64(redisConn.Do("HGET", fmt.Sprintf("Product:%s", req.ProductID), "Expire"))

		//检测商品有效期是否过期， 对彩票竞猜类的商品，有效期内才能下单
		if (expire > 0) && (expire < time.Now().UnixNano()/1e6) {
			nc.logger.Warn("商品有效期过期", zap.Int64("Expire", expire))
			errorCode = LMCError.ProductExpireError
			goto COMPLETE
		}

		// 生成订单ID
		orderID := uuid.NewV4().String()

		opk := ""

		nc.logger.Debug("GetPreKeyOrderID", zap.String("商户注册id", req.UserName), zap.String("LMCommon.VipBusinessUsername", LMCommon.VipBusinessUsername))
		if req.UserName == LMCommon.VipBusinessUsername {
			//TODO 当购买Vip会员时，不需要 opk
		} else {

			//从商户的prekeys有序集合取出一个opk
			prekeySlice, _ := redis.Strings(redisConn.Do("ZRANGE", fmt.Sprintf("prekeys:%s", req.UserName), 0, 0))
			if len(prekeySlice) > 0 {
				opk = prekeySlice[0]

				//取出后就删除此OPK
				if _, err = redisConn.Do("ZREM", fmt.Sprintf("prekeys:%s", req.UserName), opk); err != nil {
					nc.logger.Error("ZREM Error", zap.Error(err))
				}

			} else {
				nc.logger.Warn("商户的prekeys有序集合为空, 取出此商户的默认的OPK")
				opk, _ = redis.String(redisConn.Do("GET", fmt.Sprintf("DefaultOPK:%s", req.UserName)))

			}
			if opk == "" {
				nc.logger.Error("商户的OPK池是空的，并且默认OPK也是空")

				//向商户推送9-10事件通知
				go func() {
					// time.Sleep(100 * time.Millisecond)
					nc.logger.Debug("延时100ms向商户推送9-10事件通知",
						zap.String("to", req.UserName),
					)
					nc.SendOPKNoSufficientToMasterDevice(req.UserName, 0)
				}()

				errorCode = LMCError.OPKEmptyError
				goto COMPLETE
			}

			//商户的prekeys有序集合是否少于10个，如果少于，则推送报警，让SDK上传OPK
			if count, err = redis.Int(redisConn.Do("ZCOUNT", fmt.Sprintf("prekeys:%s", req.UserName), "-inf", "+inf")); err != nil {
				nc.logger.Error("ZCOUNT Error", zap.Error(err))
				errorCode = LMCError.RedisError
				goto COMPLETE
			} else {

				if count < 10 {
					nc.logger.Warn("商户的prekeys存量不足", zap.Int("count", count))

					//向商户推送9-10事件通知
					go func() {
						// time.Sleep(100 * time.Millisecond)
						nc.logger.Debug("延时100ms向商户推送9-10事件通知",
							zap.String("to", req.UserName),
						)
						nc.SendOPKNoSufficientToMasterDevice(req.UserName, count)
					}()

				} else {
					nc.logger.Debug("商户的prekeys存量", zap.Int("count", count))
				}

			}
		}

		rsp.UserName = req.UserName
		rsp.OrderType = req.OrderType
		rsp.ProductID = req.ProductID
		rsp.OrderID = orderID
		rsp.PubKey = opk

		//将订单ID保存到商户的订单有序集合orders:{username}，订单详情是 orderInfo:{订单ID}
		if _, err := redisConn.Do("ZADD", fmt.Sprintf("orders:%s", req.UserName), time.Now().UnixNano()/1e6, orderID); err != nil {
			nc.logger.Error("ZADD Error", zap.Error(err))
		}

		//订单详情
		_, err = redisConn.Do("HMSET",
			fmt.Sprintf("Order:%s", orderID),
			"OrderType", int(Global.OrderType_ORT_Normal), //订单类型是普通订单
			"BuyUser", username, //发起订单的用户id
			"BusinessUser", req.UserName, //商户的用户id
			"OrderID", orderID, //订单id
			"ProductID", req.ProductID, //商品id
			"Type", req.OrderType, //订单类型
			"State", int(Global.OrderState_OS_Undefined), //订单状态,初始为0
			"AttachHash", "", //订单内容attach的哈希值， 默认为空
			"IsPayed", LMCommon.REDISFALSE, //此订单支付状态， true- 支付完成，false-未支付
			"IsUrge", LMCommon.REDISFALSE, //催单
			"CreateAt", uint64(time.Now().UnixNano()/1e6), //毫秒
		)

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ := proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
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
9-10 OPk存量不足事件
OPk存量不足会触发此事件， 并推送给商户
*/
func (nc *NsqClient) SendOPKNoSufficientToMasterDevice(toUsername string, count int) error {
	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//查询出商户主设备
	deviceListKey := fmt.Sprintf("devices:%s", toUsername)
	deviceIDSlice, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
	for index, eDeviceID := range deviceIDSlice {
		if index == 0 {
			nc.logger.Debug("查询出商户主设备", zap.Int("index", index), zap.String("eDeviceID", eDeviceID))
			deviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
			jwtToken, _ := redis.String(redisConn.Do("GET", deviceKey))
			nc.logger.Debug("Redis GET ", zap.String("deviceKey", deviceKey), zap.String("jwtToken", jwtToken))

			//向商户主设备推送 9-10 OPK存量不足事件
			opkAlertMsg := &models.Message{}
			now := time.Now().UnixNano() / 1e6 //毫秒
			opkAlertMsg.UpdateID()

			//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
			opkAlertMsg.BuildRouter("Order", "", "Order.Frontend")
			opkAlertMsg.SetJwtToken(jwtToken)
			opkAlertMsg.SetUserName(toUsername)
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
	return nil
}

func (nc *NsqClient) DeleteAliyunOssFile(productInfo *models.ProductInfo) error {
	// New client
	client, err := oss.New(LMCommon.Endpoint, LMCommon.AccessID, LMCommon.AccessKey)
	if err != nil {
		return err
	}

	// 获取存储空间。
	bucket, err := client.Bucket(LMCommon.BucketName)
	if err != nil {
		return err
	}

	//删除文件

	if productInfo.ProductPic1Large != "" {
		err = bucket.DeleteObject(productInfo.ProductPic1Large)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic1Large:", productInfo.ProductPic1Large))
		}

	}

	if productInfo.ProductPic2Large != "" {
		err = bucket.DeleteObject(productInfo.ProductPic2Large)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic2Large:", productInfo.ProductPic2Large))
		}

	}

	if productInfo.ProductPic3Large != "" {
		err = bucket.DeleteObject(productInfo.ProductPic3Large)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("ProductPic3Large:", productInfo.ProductPic3Large))
		}

	}

	if productInfo.ShortVideo != "" {
		err = bucket.DeleteObject(productInfo.ShortVideo)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("ShortVideo:", productInfo.ShortVideo))
		}

	}

	if productInfo.DescPic1 != "" {
		err = bucket.DeleteObject(productInfo.DescPic1)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("DescPic1:", productInfo.DescPic1))
		}

	}
	if productInfo.DescPic2 != "" {
		err = bucket.DeleteObject(productInfo.DescPic2)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("DescPic2:", productInfo.DescPic2))
		}

	}
	if productInfo.DescPic3 != "" {
		err = bucket.DeleteObject(productInfo.DescPic3)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("DescPic3:", productInfo.DescPic3))
		}

	}
	if productInfo.DescPic4 != "" {
		err = bucket.DeleteObject(productInfo.DescPic4)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("DescPic4:", productInfo.DescPic4))
		}

	}
	if productInfo.DescPic5 != "" {
		err = bucket.DeleteObject(productInfo.DescPic5)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("DescPic5:", productInfo.DescPic5))
		}

	}
	if productInfo.DescPic6 != "" {
		err = bucket.DeleteObject(productInfo.DescPic6)
		if err == nil {
			nc.logger.Info("删除文件 Succeed",
				zap.String("DescPic6:", productInfo.DescPic6))
		}

	}

	return nil
}

//根据用户是否是Vip计算手续费
func (nc *NsqClient) CalculateCharge(isVip bool, orderTotalAmout float64) (float64, error) {
	if isVip {

		if orderTotalAmout < LMCommon.RateFreeAmout {
			//免手续
			return 0, nil
		} else {
			//手续减半 取小数点后两位精度
			charge := mathtool.FloatRound(orderTotalAmout*LMCommon.Rate/2, 2)
			if charge < 1 {
				charge = 1
			}
			return charge, nil
		}
	} else {
		//取小数点后两位精度
		charge := mathtool.FloatRound(orderTotalAmout*LMCommon.Rate, 2)
		if charge < 1 {
			charge = 1.0
		}
		return charge, nil
	}

}

// 9-3
/*
处理订单消息 5-1，是由ChatService转发过来的
只能是向商户下单
*/
func (nc *NsqClient) HandleOrderMsg(msg *models.Message) error {
	var err error
	errorCode := 200

	var newSeq uint64

	var isVip bool

	//经过服务端更改状态后的新的OrderProductBody字节流
	var orderProductBodyData []byte

	rsp := &Msg.SendMsgRsp{}

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
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
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		nc.logger.Debug("OrderMsg payload",
			zap.Int32("Scene", int32(req.Scene)),
			zap.Int32("Type", int32(req.Type)),
			zap.String("To", req.To), //商户账户id
			zap.String("Uuid", req.Uuid),
			zap.Uint64("SendAt", req.SendAt),
		)

		if req.To == "" {
			nc.logger.Warn("商户用户账号不能为空")
			errorCode = LMCError.BusinessUsernameIsEmptyError
			goto COMPLETE
		}
		if req.Type != Msg.MessageType_MsgType_Order {
			nc.logger.Warn("警告，不能处理非订单类型的消息")
			errorCode = LMCError.OrderMsgTypeError
			goto COMPLETE
		}

		//从redis里获取当前用户信息
		state, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "State"))

		//判断是否被封号
		if state == 2 {
			nc.logger.Warn("警告: 此用户已被封号")
			errorCode = LMCError.UserIsBlockedError
			goto COMPLETE
		} else if state == 1 {
			isVip = true
		} else {
			isVip = false
		}

		//从redis里获取目标商户的信息
		businessUserState, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", req.To), "State"))

		//判断商户是否被封号
		if businessUserState == 2 {
			nc.logger.Warn("此商户已被封号", zap.String("businessUser", req.To))
			errorCode = LMCError.BusinessUserIsBlockedError
			goto COMPLETE
		}

		//解包出 OrderProductBody
		var orderProductBody = new(Order.OrderProductBody)
		if err := proto.Unmarshal(req.Body, orderProductBody); err != nil {
			nc.logger.Warn("OrderProductBody解包错误", zap.Error(err))
			errorCode = LMCError.ProtobufUnmarshalError
			goto COMPLETE

		} else {
			//对attach进行哈希计算，以便获知订单内容是否发生改变
			attachHash := crypt.Sha1(orderProductBody.GetAttach())

			nc.logger.Debug("OrderProductBody payload",
				zap.String("OrderID", orderProductBody.OrderID),                    // 商品的订单id
				zap.String("OrderType", orderProductBody.OrderType.String()),       // 商品的订单类型
				zap.String("ProductID", orderProductBody.ProductID),                //商品id
				zap.String("BuyUser", orderProductBody.BuyUser),                    //买家
				zap.String("OpkBuyUser", orderProductBody.OpkBuyUser),              //买家的OPK公钥
				zap.String("BusinessUser", orderProductBody.BusinessUser),          //商户
				zap.String("OpkBusinessUser", orderProductBody.OpkBusinessUser),    //商户的OPK公钥
				zap.Float64("OrderTotalAmount", orderProductBody.OrderTotalAmount), //订单金额
				zap.String("Attach", orderProductBody.Attach),                      //加密的密文, hex
				zap.String("AttachHash", attachHash),                               //订单内容的哈希
				zap.ByteString("Userdata", orderProductBody.Userdata),              //透传信息 , 不加密 ，直接传过去 不处理
				zap.Int("State", int(orderProductBody.State)),                      //订单状态
			)

			//判断订单id不能为空
			if orderProductBody.OrderID == "" {
				nc.logger.Error("OrderID is empty")
				errorCode = LMCError.OrderIDIsEmptyError
				goto COMPLETE
			}

			//判断订单状态是不是 OS_Prepare, 如果是，则改为OS_SendOK
			switch Global.OrderState(orderProductBody.State) {
			case Global.OrderState_OS_Prepare:

				//总金额不能小于或等于0
				if orderProductBody.OrderTotalAmount <= 0 {
					nc.logger.Error("OrderTotalAmount is less than 0")
					errorCode = LMCError.OrderTotalAmountError
					goto COMPLETE
				}

				// 判断商品id不能为空
				if orderProductBody.ProductID == "" {
					nc.logger.Error("ProductID is empty")
					errorCode = LMCError.ProductIDIsEmptError
					goto COMPLETE
				}

				//判断买家账号id不能为空
				if orderProductBody.BuyUser == "" {
					nc.logger.Error("BuyUser is empty")
					errorCode = LMCError.BuyUserIsEmptyError
					goto COMPLETE
				}

				//判断商户的账号id不能为空
				if orderProductBody.BusinessUser == "" {
					nc.logger.Error("BusinessUser is empty")
					errorCode = LMCError.BusinessUserIsEmptyError
					goto COMPLETE
				}

				// 获取ProductID对应的商品信息
				productInfo := new(models.ProductInfo)
				if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("Product:%s", orderProductBody.ProductID))); err == nil {
					if err := redis.ScanStruct(result, productInfo); err != nil {
						nc.logger.Error("错误: ScanStruct", zap.Error(err))
						errorCode = LMCError.RedisError
						goto COMPLETE
					}
				}

				//检测商品有效期是否过期， 对彩票竞猜类的商品，有效期内才能下单
				if (productInfo.Expire > 0) && (productInfo.Expire < time.Now().UnixNano()/1e6) {
					nc.logger.Warn("商品有效期过期", zap.Int64("Expire", productInfo.Expire))
					errorCode = LMCError.ProductExpireError
					goto COMPLETE
				}

				//判断商户注册账号 是不是写死的系统商户id，如果是购买Vip, 自动接单，并返回给用户，让其发起支付操作
				if orderProductBody.BusinessUser == LMCommon.VipBusinessUsername {
					attachBytes, err := hex.DecodeString(orderProductBody.Attach) //反hex
					if err != nil {
						nc.logger.Error("orderProductBody.Attach hex.DecodeString 失败", zap.Error(err))
						errorCode = LMCError.DecodingHexError
						goto COMPLETE
					}

					attachBase, err := models.AttachBaseFromJson(attachBytes)
					if err != nil {
						nc.logger.Error("从attach里解析attachBase失败", zap.Error(err))
						errorCode = LMCError.ParseAttachError
						goto COMPLETE
					} else {
						if attachBase.BodyType == 99 {
							//从attach的Body里解析PayType， UI已经base64了，因此需要反base64
							decoded, err := base64.StdEncoding.DecodeString(attachBase.Body)
							if err != nil {
								nc.logger.Error("base64.StdEncoding.DecodeString失败", zap.Error(err))
								errorCode = LMCError.Base64DecodingError
								goto COMPLETE
							}
							vipUser, err := models.VipUserFromJson(decoded)
							if err != nil {
								nc.logger.Error("从attach的Body里解析PayType失败", zap.Error(err))
								errorCode = LMCError.ParseAttachError
								goto COMPLETE
							}
							nc.logger.Debug("购买Vip",
								zap.String("购买者", orderProductBody.BuyUser),
								zap.String("商户 ", orderProductBody.BusinessUser),
								zap.Int(" PayType ", vipUser.PayType),
							)
							//根据PayType获取到VIP价格
							vipPrice, err := nc.service.GetVipUserPrice(vipUser.PayType)
							if err != nil {
								errorCode = LMCError.QueryVipPriceError
								goto COMPLETE
							}

							//修改attachBase，将价格填入
							vipUser.Price = vipPrice.Price
							bodyTemp, _ := vipUser.ToJson()
							attachBase.Body = base64.StdEncoding.EncodeToString([]byte(bodyTemp)) // base64
							attachStr, _ := attachBase.ToJson()
							orderProductBody.AttachType = Msg.AttachType_AttachType_VipPrice
							orderProductBody.Attach = hex.EncodeToString([]byte(attachStr)) //hex

							//接单成功，当用户收到后即可发起预支付
							orderProductBody.State = Global.OrderState_OS_Taked

						} else {
							//TODO
							errorCode = LMCError.AttachBodyTypeError
							goto COMPLETE

						}

					}

				} else { // 订单下单

					//彩票类型的订单
					if Global.ProductType(productInfo.ProductType) == Global.ProductType_OT_Lottery {

						orderProductBody.State = Global.OrderState_OS_RecvOK

						orderProductBodyData, _ = proto.Marshal(orderProductBody)

						//TODO  根据Vip及订单内容生成服务费的支付数据, 并发送给买家
						//为配合flutter调试，暂时不发

						nc.SendChargeOrderIDToBuyer(req.Uuid, isVip, orderProductBody)

						//将接单转发到买家
						if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", orderProductBody.BuyUser))); err == nil {
							eRsp := &Msg.RecvMsgEventRsp{
								Scene:        Msg.MessageScene_MsgScene_S2C, //系统消息
								Type:         Msg.MessageType_MsgType_Order, //类型-订单消息
								Body:         orderProductBodyData,          //订单载体 OrderProductBody
								From:         username,                      //谁发的
								FromDeviceId: deviceID,                      //哪个设备发的
								Recv:         req.To,                        //商户账户id
								ServerMsgId:  msg.GetID(),                   //服务器分配的消息ID
								Seq:          newSeq,                        //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
								Uuid:         req.Uuid,                      //客户端分配的消息ID，SDK生成的消息id
								Time:         uint64(time.Now().UnixNano() / 1e6),
							}

							go func() {
								// time.Sleep(150 * time.Millisecond)
								nc.logger.Debug("延时150ms向买家发送彩票类型的订单, 5-2",
									zap.String("to", orderProductBody.BuyUser),
									zap.Int("State", int(orderProductBody.State)),
								)
								nc.BroadcastOrderMsgToAllDevices(eRsp, orderProductBody.BuyUser)
							}()

						}

					} else { //其它普通商品
						orderProductBody.State = Global.OrderState_OS_SendOK
						nc.logger.Debug("注意，除了预审核，其它状态都需要发送给商家", zap.Int("State", int(orderProductBody.State)))

						orderProductBodyData, _ = proto.Marshal(orderProductBody)
						//将订单转发到商户
						if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", orderProductBody.BusinessUser))); err == nil {
							eRsp := &Msg.RecvMsgEventRsp{
								Scene:        Msg.MessageScene_MsgScene_S2C, //系统消息
								Type:         Msg.MessageType_MsgType_Order, //类型-订单消息
								Body:         orderProductBodyData,          //订单载体 OrderProductBody
								From:         username,                      //谁发的
								FromDeviceId: deviceID,                      //哪个设备发的
								Recv:         req.To,                        //商户账户id
								ServerMsgId:  msg.GetID(),                   //服务器分配的消息ID
								Seq:          newSeq,                        //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
								Uuid:         req.Uuid,                      //客户端分配的消息ID，SDK生成的消息id
								Time:         uint64(time.Now().UnixNano() / 1e6),
							}

							go func() {
								// time.Sleep(150 * time.Millisecond)
								nc.logger.Debug("延时150ms向商家发送非彩票普通商品的订单消息, 5-2",
									zap.String("to", orderProductBody.BuyUser),
									zap.Int("State", int(orderProductBody.State)),
								)
								nc.BroadcastOrderMsgToAllDevices(eRsp, orderProductBody.BusinessUser)
							}()

						}
					}

					//对attach进行哈希计算，以便获知订单内容是否发生改变
					attachHash := crypt.Sha1(orderProductBody.Attach)

					// 将订单信息OrderProductBody数据缓存在redis里的一个哈希表里, 以 ServerMsgId 对应, 在消息ack时用到
					orderProductBodyKey := fmt.Sprintf("OrderProductBody:%s", msg.GetID())
					_, err = redisConn.Do("HMSET",
						orderProductBodyKey,
						"Username", username,
						"OrderID", orderProductBody.OrderID,
						"OrderType", int(orderProductBody.OrderType),
						"ProductID", orderProductBody.ProductID,
						"BuyUser", orderProductBody.BuyUser,
						"OpkBuyUser", orderProductBody.OpkBuyUser,
						"BusinessUser", orderProductBody.BusinessUser,
						"OpkBusinessUser", orderProductBody.OpkBusinessUser,
						"OrderTotalAmount", orderProductBody.OrderTotalAmount, //订单金额
						"AttachType", int(orderProductBody.AttachType), //附件类型
						"Attach", orderProductBody.Attach, //订单内容，UI负责构造
						"AttachHash", attachHash, //订单内容的哈希值
						"UserData", orderProductBody.Userdata, //透传数据
						"State", orderProductBody.State, //订单状态
					)

				}

			default:
				nc.logger.Error("订单状态未定义", zap.Error(err))
				errorCode = LMCError.UnkownOrderTypeError
				goto COMPLETE
			}
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
				Uuid:        req.Uuid, // SDK构造，用来识别那条消息
				ServerMsgId: msg.GetID(),
				Seq:         curSeq,
				Time:        uint64(time.Now().UnixNano() / 1e6), //毫秒
			}
			data, _ := proto.Marshal(rsp)
			msg.FillBody(data)
		}
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("HandleOrderMsg 5-1: Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("HandleOrderMsg 5-1: Failed to send  message to ProduceChannel", zap.Error(err))
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

	var newSeq uint64

	var orderBodyData []byte
	var orderID, productID string
	var orderType int //订单类型

	var buyUser, businessUser string
	var toUsername string //目标用户账号id，可能是商户，可能是买家

	var attachHash string
	var orderTotalAmount float64 //订单金额
	var isPayed, isUrge bool
	var orderIDKey string

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //当前用户
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
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {

		if req.OrderBody.OrderID == "" {
			nc.logger.Error("OrderID is empty")
			errorCode = LMCError.OrderIDIsEmptyError
			goto COMPLETE
		}
		orderID = req.OrderBody.OrderID
		orderType = int(req.OrderBody.OrderType)

		//根据订单id获取buyUser及businessUser是谁
		orderIDKey = fmt.Sprintf("Order:%s", orderID)
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", orderIDKey)); err != nil {
			nc.logger.Error("EXISTS 错误", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		} else {
			if isExists == false {
				nc.logger.Error("orderID is not exists")
				errorCode = LMCError.OrderIDIsNotExistsError
				goto COMPLETE
			}
		}

		//获取订单的具体信息
		isPayed, _ = redis.Bool(redisConn.Do("HGET", orderIDKey, "IsPayed"))
		isUrge, _ = redis.Bool(redisConn.Do("HGET", orderIDKey, "IsUrge"))
		productID, _ = redis.String(redisConn.Do("HGET", orderIDKey, "ProductID"))
		buyUser, _ = redis.String(redisConn.Do("HGET", orderIDKey, "BuyUser"))
		businessUser, _ = redis.String(redisConn.Do("HGET", orderIDKey, "BusinessUser"))
		orderTotalAmount, _ = redis.Float64(redisConn.Do("HGET", orderIDKey, "OrderTotalAmount"))
		attachHash, _ = redis.String(redisConn.Do("HGET", orderIDKey, "AttachHash"))

		// if err != nil {
		// 	nc.logger.Error("从Redis里取出此 Order 对应的businessUser Error", zap.String("orderIDKey", orderIDKey), zap.Error(err))
		// }

		if productID == "" {
			nc.logger.Error("ProductID is empty")
			errorCode = LMCError.ProductIDIsEmptError
			goto COMPLETE
		}

		if buyUser == "" {
			nc.logger.Error("BuyUser is empty")
			errorCode = LMCError.BuyUserIsEmptyError
			goto COMPLETE
		}

		if businessUser == "" {
			nc.logger.Error("BusinessUse is empty")
			errorCode = LMCError.BusinessUserIsEmptyError
			goto COMPLETE
		}

		if orderTotalAmount <= 0 {
			nc.logger.Error("OrderTotalAmount is less than 0")
			errorCode = LMCError.OrderTotalAmountError
			goto COMPLETE
		}

		//明确目标用户是买家还是商户
		if buyUser == username {
			toUsername = businessUser
		} else {
			toUsername = buyUser
		}

		nc.logger.Debug("ChangeOrderState",
			zap.Int("State", int(req.State)), //需要更新的状态
			zap.Uint64("TimeAt", req.TimeAt),
			zap.String("OrderID", orderID),
			zap.Int("OrderType", int(orderType)),
			zap.String("ProductID", productID),
			zap.String("BuyUser", buyUser),
			zap.String("BusinessUser", businessUser),
			zap.String("AttachHash", attachHash),
			zap.String("当前操作者账号 username", username),
			zap.String("目标用户账号 toUsername", toUsername),
			zap.Float64("OrderTotalAmount", orderTotalAmount),
		)

		//从redis里获取买家信息
		buyerState, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", buyUser), "State"))

		//判断用户是否被封号
		if buyerState == 2 {
			nc.logger.Warn("此用户已被封号", zap.String("User", buyUser))
			errorCode = LMCError.UserIsBlockedError
			goto COMPLETE
		}

		//从redis里获取商户的信息
		businessUserState, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", businessUser), "State"))

		//判断商户是否被封号
		if businessUserState == 2 {
			nc.logger.Warn("f", zap.String("businessUser", businessUser))
			errorCode = LMCError.BusinessUserIsBlockedError
			goto COMPLETE
		}

		//判断此订单是否已经在商户的有序集合里orders:{账号id}
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("orders:%s", businessUser), orderID); err == nil {
			if reply == nil {
				//此订单id不属于此商户
				nc.logger.Error("此订单id不属于此商户",
					zap.String("OrderID", orderID),
				)
				errorCode = LMCError.OrderIDNotBelongToError
				goto COMPLETE
			}

		}

		//获取当前订单的状态
		curState, err := redis.Int(redisConn.Do("HGET", orderIDKey, "State"))

		//根据当前订单的状态做逻辑，某些状态不能更新
		switch Global.OrderState(curState) {

		case Global.OrderState_OS_Undefined: //未定义
			//将redis里的订单信息哈希表状态字段设置为最新状态
			_, err = redisConn.Do("HSET", orderIDKey, "State", int(req.State))

		case Global.OrderState_OS_Done: //当前处于: 完成订单
			if businessUser == username {
				nc.logger.Warn("警告: 当前状态处于完成订单状态, 不能更改为其它")
				errorCode = LMCError.OrderStatusBusinessChangeError
				goto COMPLETE
			} else {
				if req.State != Global.OrderState_OS_Confirm {
					nc.logger.Warn("警告: 当前状态处于完成订单状态, 只能选择确认")
					errorCode = LMCError.OrderStatusChangeConfirmError
					goto COMPLETE
				}
			}

		case Global.OrderState_OS_ApplyCancel: //当前处于: 买家申请撤单

			// if businessUser == username { //只有商户才能有权更改订单状态为完成、撤单、更改订单内容（金额）
			// 	if req.State== Global.OrderState_OS_Done ||
			// 		req.State== Global.OrderState_OS_Cancel {
			// 		//pass
			// 	} else {
			// 		nc.logger.Warn("警告: 只有商户才能有权更改订单状态为撤单或完成")
			// 		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			// 		goto COMPLETE
			// 	}
			// }

		case Global.OrderState_OS_Confirm: //当前处于:  确认收货
			nc.logger.Warn("警告: 此订单已经确认收货,不能再更改其状态")
			errorCode = LMCError.OrderStatusConfirmIsDoneError
			goto COMPLETE

		case Global.OrderState_OS_Cancel: //当前处于: 撤单
			nc.logger.Warn("警告: 此订单已撤单,不能再更改其状态")
			errorCode = LMCError.OrderStatusIsCancelError
			goto COMPLETE

		case Global.OrderState_OS_AttachChange: //当前处于: 订单内容发生更改

		case Global.OrderState_OS_Paying: //当前状态为支付中， 锁定订单，不能更改，直到支付完成
			nc.logger.Warn("警告: 此订单当前状态为支付中, 不能更改状态")
			errorCode = LMCError.OrderStatusIsPayingError
			goto COMPLETE

		case Global.OrderState_OS_IsPayed: // 已支付， 支付成功
			nc.logger.Debug("此订单当前状态为已支付， 支付成功")

		// case Global.OrderState_OS_Overdue: // 已逾期, 订单终止，无法再更改状态了的
		// 	nc.logger.Warn("警告: 此订单当前状态为已逾期, 不能更改状态")
		// 	errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		// 	goto COMPLETE

		case Global.OrderState_OS_Refuse: //当前处于: 已拒单
			nc.logger.Warn("警告: 此订单已拒单,不能再更改其状态")
			errorCode = LMCError.OrderStatusIsRefusedError
			goto COMPLETE

		case Global.OrderState_OS_Urge: // 买家催单, 商户可以回复7， 只能催一次
			if req.State == Global.OrderState_OS_Urge {
				nc.logger.Warn("警告: 此订单当前状态为买家催单, 只能催一次")

				errorCode = LMCError.OrderStatusIsUrgedError
				goto COMPLETE

			}

		}

		//根据最新状态要处理的逻辑
		switch req.State {
		case Global.OrderState_OS_Taked: //已接单, 向买家推送通知
			if toUsername == buyUser {
				nc.logger.Warn("警告: 买家不能接单，否则就是 SDK逻辑错误")

				errorCode = LMCError.OrderStatusIsBuyerError
				goto COMPLETE
			}
			nc.logger.Debug("已接单, 向买家推送通知", zap.String("BusinessUser", businessUser))

			//通知买家
			orderBodyData, _ = proto.Marshal(&Order.OrderProductBody{
				OrderID:      orderID,
				OrderType:    Global.OrderType(int32(orderType)), //订单类型
				ProductID:    productID,
				BuyUser:      buyUser,
				BusinessUser: businessUser,
				State:        Global.OrderState_OS_Taked,
			})

		case Global.OrderState_OS_AttachChange: //订单内容发生更改, 需要解包
			if isPayed {
				nc.logger.Error("完成支付之后不能修改订单内容及金额")
				errorCode = LMCError.OrderStatusPayedError
				goto COMPLETE
			}

			//判断订单id不能为空
			if req.OrderBody.OrderID == "" {
				nc.logger.Error("OrderID is empty")
				errorCode = LMCError.OrderIDIsEmptyError
				goto COMPLETE
			}

			if req.OrderBody.GetOrderTotalAmount() <= 0 {
				nc.logger.Error("OrderTotalAmount is less than  0")
				errorCode = LMCError.OrderTotalAmountError
				goto COMPLETE
			}

			//订单金额改变或者哈希发生改变
			cur_attachHash := crypt.Sha1(req.OrderBody.GetAttach())
			if orderTotalAmount != req.OrderBody.GetOrderTotalAmount() || attachHash != cur_attachHash {
				nc.logger.Debug("OrderBody change，订单内容或金额发生改变",
					zap.String("OrderID", req.OrderBody.OrderID),
					zap.String("OrderTyp[e", req.OrderBody.OrderType.String()),
					zap.String("ProductID", req.OrderBody.GetProductID()),
					zap.String("BuyUser", req.OrderBody.GetBuyUser()),
					zap.String("OpkBuyUser", req.OrderBody.GetOpkBuyUser()),
					zap.String("BusinessUser", req.OrderBody.BusinessUser),
					zap.String("OpkBusinessUser", req.OrderBody.GetOpkBusinessUser()),
					zap.Float64("OrderTotalAmount", req.OrderBody.GetOrderTotalAmount()),
					zap.String("Attach", req.OrderBody.GetAttach()), //加密的密文 hex
					zap.String("AttachHash", cur_attachHash),        //订单内容哈希
					zap.Int("State", int(req.OrderBody.GetState())), //订单状态
				)

				//将OrderTotalAmount, Attach， AttachHash， UserData 更新到redis里
				_, err = redisConn.Do("HSET", orderIDKey, "OrderTotalAmount", req.OrderBody.GetOrderTotalAmount())
				_, err = redisConn.Do("HSET", orderIDKey, "Attach", req.OrderBody.GetAttach())
				_, err = redisConn.Do("HSET", orderIDKey, "AttachHash", cur_attachHash)
				_, err = redisConn.Do("HSET", orderIDKey, "UserData", req.OrderBody.GetUserdata())
			}

			//通知对方
			orderBodyData, _ = proto.Marshal(req.OrderBody)

		case Global.OrderState_OS_Done: //完成订单, 商户发送的
			if isPayed == false {
				nc.logger.Error("完成订单, 商户发送的， 但是未完成支付 ")
				errorCode = LMCError.OrderStatusNotPayError
				goto COMPLETE
			}

			//通知买家
			orderBodyData, _ = proto.Marshal(&Order.OrderProductBody{
				OrderID:      orderID,
				OrderType:    Global.OrderType(int32(orderType)), //订单类型
				ProductID:    productID,
				BuyUser:      buyUser,
				BusinessUser: businessUser,
				State:        Global.OrderState_OS_Done,
			})

		case Global.OrderState_OS_ApplyCancel: // 买家申请撤单
			if isPayed == false {
				nc.logger.Debug("买家申请撤单, 但是未完成支付")
			}

			//通知商家
			orderBodyData, _ = proto.Marshal(&Order.OrderProductBody{
				OrderID:      orderID,
				OrderType:    Global.OrderType(int32(orderType)), //订单类型
				ProductID:    productID,
				BuyUser:      buyUser,
				BusinessUser: businessUser,
				State:        Global.OrderState_OS_ApplyCancel,
			})

		case Global.OrderState_OS_Cancel: // 商户同意撤单
			if isPayed == false {
				nc.logger.Debug("商户同意撤单, 但买家未完成支付")
			} else {
				//TODO 扣除手续费
				nc.logger.Debug("商户同意撤单, 买家完成支付, 需要退款")
				//向钱包服务端发送一条grpc转账消息，将连米代币从中间账号转到买家的钱包， 实现退款
				ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
				transferResp, err := nc.service.TransferByOrder(ctx, &Wallet.TransferReq{
					OrderID: orderID,
					PayType: LMCommon.OrderTransferForCancel, //退款
				})
				if err != nil {
					nc.logger.Error("walletSvc.TransferByOrder Error", zap.Error(err))
					errorCode = LMCError.WalletTranferError

					goto COMPLETE
				} else {
					nc.logger.Debug("walletSvc.TransferByOrder succeed", zap.Int32("ErrCode", transferResp.ErrCode), zap.String("ErrMsg", transferResp.ErrMsg))

				}
			}

			//通知买家
			orderBodyData, _ = proto.Marshal(&Order.OrderProductBody{
				OrderID:      orderID,
				OrderType:    Global.OrderType(int32(orderType)), //订单类型
				ProductID:    productID,
				BuyUser:      buyUser,
				BusinessUser: businessUser,
				State:        Global.OrderState_OS_Cancel,
			})

		case Global.OrderState_OS_Processing: //订单处理中，一般用于商户，安抚下单的
			//通知买家
			orderBodyData, _ = proto.Marshal(&Order.OrderProductBody{
				OrderID:      orderID,
				OrderType:    Global.OrderType(int32(orderType)), //订单类型
				ProductID:    productID,
				BuyUser:      buyUser,
				BusinessUser: businessUser,
				State:        Global.OrderState_OS_Processing,
			})

		case Global.OrderState_OS_Confirm: //确认收货
			if isPayed == false {
				nc.logger.Error("买家确认收货, 但是未完成支付")
				errorCode = LMCError.OrderStatusNotPayError
				goto COMPLETE
			}

			//如果当前状态不是已完成 ，买家不能确认收货
			if Global.OrderState(curState) != Global.OrderState_OS_Done {
				nc.logger.Error("买家确认收货, 但是此订单未完成")
				errorCode = LMCError.OrderStatusNotDoneError
				goto COMPLETE
			}

			//向钱包服务端发送一条转账grpc消息，将连米代币从中间账号转到商户的钱包
			ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
			transferResp, err := nc.service.TransferByOrder(ctx, &Wallet.TransferReq{
				OrderID: orderID,
				PayType: LMCommon.OrderTransferForDone, //完成结算
			})
			if err != nil {
				nc.logger.Error("walletSvc.TransferByOrder Error", zap.Error(err))
				errorCode = LMCError.WalletTranferError
				goto COMPLETE
			} else {
				nc.logger.Debug("walletSvc.TransferByOrder succeed", zap.Int32("ErrCode", transferResp.ErrCode), zap.String("ErrMsg", transferResp.ErrMsg))

			}

			//通知商户
			orderBodyData, _ = proto.Marshal(&Order.OrderProductBody{
				OrderID:      orderID,
				OrderType:    Global.OrderType(int32(orderType)), //订单类型
				ProductID:    productID,
				BuyUser:      buyUser,
				BusinessUser: businessUser,
				State:        Global.OrderState_OS_Confirm,
			})

		case Global.OrderState_OS_Paying: // 支付中， 此状态不能由用户设置
			nc.logger.Error("此状态不能由用户设置为支付中")
			errorCode = LMCError.OrderStatusCannotChangetoPayingError
			goto COMPLETE

		case Global.OrderState_OS_Overdue: //已逾期

		case Global.OrderState_OS_IsPayed: //已支付， 支付成功
			if isPayed == false {
				nc.logger.Error("未完成支付之前不能设置为已支付")
				errorCode = LMCError.OrderStatusNotPayError
				goto COMPLETE
			}

		case Global.OrderState_OS_Refuse: //商户拒单， 跟已接单是相反的操作
			if isPayed == false {
				nc.logger.Debug("商户拒单, 但买家未完成支付")
				errorCode = LMCError.OrderStatusNotPayError
				goto COMPLETE
			}

			//向买家发送通知，并且发送grpc消息给钱包服务端，完成退款操作
			ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
			transferResp, err := nc.service.TransferByOrder(ctx, &Wallet.TransferReq{
				OrderID: orderID,
				PayType: LMCommon.OrderTransferForCancel, //退款
			})

			if err != nil {
				nc.logger.Error("walletSvc.TransferByOrder Error", zap.Error(err))
				errorCode = LMCError.WalletTranferError
				goto COMPLETE
			} else {
				nc.logger.Debug("walletSvc.TransferByOrder succeed", zap.Int32("ErrCode", transferResp.ErrCode), zap.String("ErrMsg", transferResp.ErrMsg))

			}

			//通知买家
			orderBodyData, _ = proto.Marshal(&Order.OrderProductBody{
				OrderID:      orderID,
				OrderType:    Global.OrderType(int32(orderType)), //订单类型
				ProductID:    productID,
				BuyUser:      buyUser,
				BusinessUser: businessUser,
				State:        Global.OrderState_OS_Refuse,
			})

		case Global.OrderState_OS_Urge:
			if isPayed == false {
				nc.logger.Error("买家催单, 但是未完成支付")
				errorCode = LMCError.OrderStatusNotPayError
				goto COMPLETE
			}
			if isUrge == true {
				nc.logger.Error("买家催单, 只能催一次")
				errorCode = LMCError.OrderStatusOnceUrgedError
				goto COMPLETE
			}

			//通知商户
			orderBodyData, _ = proto.Marshal(&Order.OrderProductBody{
				OrderID:      orderID,
				OrderType:    Global.OrderType(int32(orderType)), //订单类型
				ProductID:    productID,
				BuyUser:      buyUser,
				BusinessUser: businessUser,
				State:        Global.OrderState_OS_Urge,
			})

		case Global.OrderState_OS_Expedited:
			if buyerState != 1 {
				nc.logger.Error("买家加急, 但不是VIP用户")
				errorCode = LMCError.OrderStatusVipExpeditedError
				goto COMPLETE
			}
			if isPayed == false {
				nc.logger.Error("买家加急, 但是未完成支付")
				errorCode = LMCError.OrderStatusNotPayError
				goto COMPLETE
			}

			//通知商户
			orderBodyData, _ = proto.Marshal(&Order.OrderProductBody{
				OrderID:      orderID,
				OrderType:    Global.OrderType(int32(orderType)), //订单类型
				ProductID:    productID,
				BuyUser:      buyUser,
				BusinessUser: businessUser,
				State:        Global.OrderState_OS_Urge,
			})

		default:
			nc.logger.Debug("Do nothing ...'", zap.Int("State", int(req.State)))
		}

		//将redis里的订单信息哈希表状态字段设置为最新状态
		_, err = redisConn.Do("HSET", orderIDKey, "State", int(req.State))

		if int(req.State) > 1 {
			//将最新订单状态转发到目标用户
			if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", toUsername))); err != nil {
				nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				errorCode = LMCError.RedisError
				goto COMPLETE
			}

			eRsp := &Msg.RecvMsgEventRsp{
				Scene:        Msg.MessageScene_MsgScene_S2C, //系统消息
				Type:         Msg.MessageType_MsgType_Order, //类型-订单消息
				Body:         orderBodyData,                 //发起方的body负载
				From:         username,                      //谁发的
				FromDeviceId: deviceID,                      //哪个设备发的
				ServerMsgId:  msg.GetID(),                   //服务器分配的消息ID
				Seq:          newSeq,                        //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
				// Uuid:         req.Uuid,                   //客户端分配的消息ID，SDK生成的消息id
				Time: uint64(time.Now().UnixNano() / 1e6),
			}

			//向目标用户发送订单消息状态的更改
			go func() {
				// time.Sleep(100 * time.Millisecond)
				nc.logger.Debug("延时100ms向目标用户发送订单消息状态的更改, 5-2",
					zap.String("to", toUsername),
					zap.String("状态改变", fmt.Sprintf("目标用户(%s), (%d)->(%d)", toUsername, curState, req.State)),
				)
				nc.BroadcastOrderMsgToAllDevices(eRsp, toUsername)
			}()

		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
		msg.FillBody(nil)
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
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

	var count int

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
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
	userKey := fmt.Sprintf("userData:%s", username)
	userType, _ := redis.Int(redisConn.Do("HGET", userKey, "UserType"))
	if userType != 2 {
		nc.logger.Error("只有商户才能查询OPK存量")
		errorCode = LMCError.PreKeyGetCountError
		goto COMPLETE
	}

	if count, err = redis.Int(redisConn.Do("ZCOUNT", fmt.Sprintf("prekeys:%s", username), "-inf", "+inf")); err != nil {
		nc.logger.Error("ZCOUNT Error", zap.Error(err))
		errorCode = LMCError.RedisError

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
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
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
		// targetMsg.SetTaskID(uint32(taskId))
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
			zap.String("DeviceID:", eDeviceID),
			zap.Int64("Now", time.Now().UnixNano()/1e6))

	}

	return nil
}

/*
向目标用户账号的所有端推送消息， 接收端会触发接收消息事件
业务号:  BusinessType_Msg(5)
业务子号:  MsgSubType_RecvMsgEvent(2)
*/
func (nc *NsqClient) BroadcastOrderMsgToAllDevices(rsp *Msg.RecvMsgEventRsp, toUsername string) error {
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

	//订单消息具体内容
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
		targetMsg.BuildHeader("OrderService", time.Now().UnixNano()/1e6)

		targetMsg.UpdateID()
		//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
		targetMsg.BuildRouter("Order", "", "Order.Frontend")

		targetMsg.SetJwtToken(curJwtToken)
		targetMsg.SetUserName(toUsername)
		targetMsg.SetDeviceID(eDeviceID)
		// opkAlertMsg.SetTaskID(uint32(taskId))
		targetMsg.SetBusinessTypeName("Order")
		targetMsg.SetBusinessType(uint32(Global.BusinessType_Msg))           //消息模块 5
		targetMsg.SetBusinessSubType(uint32(Global.MsgSubType_RecvMsgEvent)) //接收消息事件 2

		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

		targetMsg.SetCode(200) //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Order.Frontend"
		rawData, _ := json.Marshal(targetMsg)
		if err := nc.Producer.Public(topic, rawData); err == nil {
			nc.logger.Info("5-2 Message succeed send to dispatcher", zap.String("topic", topic))
		} else {
			nc.logger.Error("Failed to send 5-2 message to dispatcher", zap.Error(err))
		}

		nc.logger.Info("Broadcast Msg To All Devices Succeed",
			zap.String("Username:", toUsername),
			zap.String("DeviceID:", eDeviceID),
			zap.Int64("Now", time.Now().UnixNano()/1e6))

		_ = err

	}

	return nil
}
