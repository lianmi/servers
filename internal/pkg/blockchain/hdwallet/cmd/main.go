package main

import(
	"log"
	"github.com/lianmi/servers/internal/pkg/blockchain/hdwallet"
)


func main() {
	mnemonic, err := hdwallet.NewMnemonic(128)
	if err != nil {
		log.Println(err)
		return;
	} else {
		log.Println(mnemonic)
	}
}