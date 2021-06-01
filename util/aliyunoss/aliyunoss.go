package aliyunoss

import (
	"crypto/md5"
	"encoding/hex"
	"io"

	// "log"
	"os"
	"path"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	LMCommon "github.com/lianmi/servers/internal/common"
)

func UploadOssFile(modName, userName, localFileName string) (string, error) {
	var err error
	var objectName string

	// <yourObjectName>上传文件到OSS时需要指定包含文件后缀在内的完整路径，例如abc/efg/123.jpg。
	// 阿里云会自动创建各级子目录

	f, err := os.Open(localFileName)
	if err != nil {
		// log.Println("Error: ", err)
		return "", err
	}

	defer f.Close()

	md5hash := md5.New()
	if _, err := io.Copy(md5hash, f); err != nil {
		// log.Println("Copy", err)
		return "", err
	}

	md5hash.Sum(nil)
	// log.Printf("%x\n", md5hash.Sum(nil))

	md5Str := hex.EncodeToString(md5hash.Sum(nil))
	// log.Printf("md5: %s\n", md5Str)

	//上传的文件名： md5 +  原来的后缀名
	fileExt := path.Ext(localFileName)
	objectName = modName + "/" + userName + "/" + time.Now().Format("2006/01/02/") + md5Str + fileExt

	// 创建OSSClient实例。
	client, err := oss.New(LMCommon.Endpoint, LMCommon.SuperAccessID, LMCommon.SuperAccessKeySecret)
	if err != nil {
		// log.Println("oss Error:", err)
		return "", err

	} else {
		// OSS操作。
		// log.Println("利用临时STS创建OSSClient实例 ok")
	}

	// 获取存储空间。
	bucket, err := client.Bucket(LMCommon.BucketName)
	if err != nil {
		return "", err
	}

	// 上传文件。
	// log.Println("objectName... ", objectName)
	// log.Println("上传文件... ", localFileName)
	err = bucket.PutObjectFromFile(objectName, localFileName)
	if err != nil {
		return "", err
	} else {
		// log.Println("上传完成", objectName)
	}

	return objectName, nil

}
