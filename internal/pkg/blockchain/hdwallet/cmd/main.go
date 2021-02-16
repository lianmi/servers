package main

import (
	"log"

	"github.com/lianmi/servers/internal/pkg/blockchain/hdwallet"
)

func main() {
	mnemonic, err := hdwallet.NewMnemonic(128)
	if err != nil {
		log.Println(err)
		return
	} else {
		log.Println(mnemonic)
	}

	wallet, err := hdwallet.NewFromMnemonic(mnemonic, "")
	if err != nil {
		log.Println(err)
		return
	}
	seed, err := hdwallet.NewSeedFromMnemonic(mnemonic, "")
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("seed length: ", len(seed))

	// path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	// account, err := wallet.Derive(path, true)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// publicKeyBase64, _ := wallet.PublicKeyToBase64(account)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// log.Println("m/44'/60'/0'/0/0", publicKeyBase64)

	// publicKeyBytes, _ := wallet.PublicKeyFromBase64(publicKeyBase64)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// log.Println("publicKeyBytes length: ", len(publicKeyBytes))

	_ = wallet
}
