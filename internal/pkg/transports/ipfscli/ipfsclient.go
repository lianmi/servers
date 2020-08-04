package ipfscli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
)

const (
	LocalHost = "http://127.0.0.1:11080"
)

func New() {
	// Where your local node is running on localhost:5001
	sh := shell.NewShell("127.0.0.1:5001")

	//上传文字
	cid, err := sh.Add(strings.NewReader("hello world!"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("added %s\n", cid)

	//生成访问的URL
	fmt.Printf("url: %s\n", LocalHost+"/"+cid)

	//Get, 第二个参数是本地存放的文件名
	if err := sh.Get(cid, "./download/h.txt"); err != nil {
		fmt.Printf("Get error: %s", err.Error())
		return
	}

	//上传本地图片
	picFile := "/Users/mac/developments/lianmi/ipfs/pics/shuangseq001.jpg"
	content, err := ioutil.ReadFile(picFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	picId, err := sh.Add(strings.NewReader(string(content)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("added a picfile %s\n", picId)

	//生成访问的URL
	fmt.Printf("pic url: %s\n", LocalHost+"/01.jpg")

	//Get, 第二个参数是本地存放的文件名
	if err := sh.Get(picId, "./download/01.jpg"); err != nil {
		fmt.Printf("Get error: %s", err.Error())
		return
	}
}
