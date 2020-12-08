/*
上传 图片文件到阿里云
第一步:  ./lmSdkClient oss osstoken
第二步:  ./lmSdkClient oss upload -f ~/Downloads/shuangseqiu.jpeg -b lianmi-ipfs -d products
*/
package cmd

import (
	"log"

	"crypto/md5"
	"encoding/hex"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path"
)

func handleError(err error) {
	log.Println("Error:", err)
	os.Exit(-1)
}

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("upload called")
		localFileName, _ := cmd.PersistentFlags().GetString("file")
		if localFileName == "" {
			log.Println("Error:  localFileName is empty")
			return
		}
		// bucketName := "lianmi-images" //公共读

		bucketName, _ := cmd.PersistentFlags().GetString("bucket")
		if bucketName == "" {
			log.Println("Error:  bucket name is empty")
			return
		}

		dir, _ := cmd.PersistentFlags().GetString("dir")
		if dir == "" {
			log.Println("Error:  dir name is empty")
			return
		}

		redisConn, err := redis.Dial("tcp", "127.0.0.1:6379")
		if err != nil {
			log.Fatalln(err)
			return
		}

		defer redisConn.Close()

		endpoint := "https://oss-cn-hangzhou.aliyuncs.com"

		accessKeyID, _ := redis.String(redisConn.Do("GET", "OSSAccessKeyId"))
		accessSecretKey, _ := redis.String(redisConn.Do("GET", "OSSAccessKeySecret"))
		securityToken, _ := redis.String(redisConn.Do("GET", "OSSSecurityToken"))

		log.Println("endpoint: ", endpoint)
		log.Println("accessKeyID: ", accessKeyID)
		log.Println("accessSecretKey: ", accessSecretKey)
		if endpoint == "" || accessKeyID == "" || accessSecretKey == "" {
			log.Printf("endpoint accessKeyID accessSecretKey is empty \n")
			return
		}

		// 创建OSSClient实例。
		client, err := oss.New(endpoint, accessKeyID, accessSecretKey, oss.SecurityToken(securityToken))
		if err != nil {
			log.Println("Error:", err)
			return

		} else {
			// OSS操作。
			log.Println("利用临时STS创建OSSClient实例 ok")
		}
		// <yourObjectName>上传文件到OSS时需要指定包含文件后缀在内的完整路径，例如abc/efg/123.jpg。
		// 阿里云会自动创建各级子目录
		// localFileName := "./cat.jpg"

		f, err := os.Open(localFileName)
		if err != nil {
			log.Println("Error: ", err)
			return
		}

		defer f.Close()

		md5hash := md5.New()
		if _, err := io.Copy(md5hash, f); err != nil {
			log.Println("Copy", err)
			return
		}

		md5hash.Sum(nil)
		// fmt.Printf("%x\n", md5hash.Sum(nil))

		md5Str := hex.EncodeToString(md5hash.Sum(nil))
		log.Printf("md5: %s\n", md5Str)

		//上传的文件名： md5 +  原来的后缀名
		fileExt := path.Ext(localFileName)
		// objectName := "generalproduct/" + time.Now().Format("2006/01/02/") + userName + "/" + md5Str + fileExt
		objectName := dir + "/" + md5Str + fileExt
		log.Printf("objectName: %s\n", objectName)

		// 获取存储空间。
		bucket, err := client.Bucket(bucketName)
		if err != nil {
			handleError(err)
		}
		// 上传文件。
		err = bucket.PutObjectFromFile(objectName, localFileName)
		if err != nil {
			handleError(err)
		} else {
			url := "https://" + bucketName + ".oss-cn-hangzhou.aliyuncs.com/" + objectName
			log.Println("上传完成, url: ", url)
		}

	},
}

func init() {
	//子命令
	ossCmd.AddCommand(uploadCmd)

	uploadCmd.PersistentFlags().StringP("file", "f", "", "local flie (full path) ")  //本地文件
	uploadCmd.PersistentFlags().StringP("bucket", "b", "lianmi-ipfs", "bucket name") //bucket
	uploadCmd.PersistentFlags().StringP("dir", "d", "", "dir")                       //bucket里的存放目录
}
