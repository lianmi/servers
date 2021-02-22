package main

import (
	"fmt"

	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/util/array"
)

func main() {
	size := len(LMCommon.OSSUploadPicPrefix)
	productPic1Large := "https://lianmi-ipfs.oss-cn-hangzhou.aliyuncs.com/products/id58/2021/02/21/90e1780e70906408314ba917e206bba3"
	productPic1Large = array.Substr2(productPic1Large, size, len(productPic1Large))
	fmt.Println(productPic1Large)
}
