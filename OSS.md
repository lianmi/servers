#  阿里云OSS使用规范
## 一、说明

   ### 1. 上传:  
      都必须经过SDK进行上传, 返回objid

   ### 2. UI层访问都是用URL方式， 无须oss的token 
   #### 例子1：
    用户头像 https://lianmi-ipfs.oss-cn-hangzhou.aliyuncs.com/avatars/3bd9492cb75f990b3effd5b39614f510.jpeg?x-oss-process=image/resize,w_50/quality,q_50

   #### 例子2：
    商品图片 https://lianmi-ipfs.oss-cn-hangzhou.aliyuncs.com/product/215b66d14111da360261206e348c3223.jpg

    
   ### 3. UI层要访问私有的图片 
   例如： 身份证，营业执照，订单上链，这些不提供URL，需要SDK下载，并向UI返回缓存目录里的具体文件路径
   营业执照下载成功 例子:  /data/data/cache/stores/%E8%90%A5%E4%B8%9A%E6%89%A7%E7%85%A7.jpg

## 二、约定bucket 是唯一一个 
```
lianmi-ipfs
```

## 三、 目录规范

### 1、 系统图片目录，包括通用商品， 默认头像    
####   (1). 权限 
     用户只能匿名读(read-only)， 由后台超级管理员上传(CRUD)

####   (2). 目录约定
    通用商品: generalproduct/xxx.jpg
    头像:   generalavatars/xxx.jpg

### 2、 店铺(形象图片，营业执照)、商品、头像、群头像的图片及短视频
 
####    (1). 权限 
     用户可以上传，但不能删除。其它用户只能匿名读

####    (2). 目录约定 
    店铺:   stores/id1/2020-12-04/xxx.jpg
    商品:   products/id1/2020-12-04/xxx.jpg
    用户头像:   avatars/id1/2020-12-04/xxx.jpg
    群头像:   teamicons/id1/2020-12-04/xxx.jpg

### 3、 实名用户的身份证照片 ，订单上链照片 、聊天消息里的图片及短视频
 
####    (1). 权限 
     用户可以上传，但不能删除。其它用户不能读, 需要SDK来读
     保证每个App用户之间的数据隔离(暂时不做隔离)

####    (2). 目录约定 
    用户身份证照片:   users/id1/2020-12-04/身份证.jpg
    订单上链照片:   orders/id1/2020-12-04/xxx.jpg
    聊天消息里的图片及短视频:   msg/id1/2020-12-04/xxx.jpg
