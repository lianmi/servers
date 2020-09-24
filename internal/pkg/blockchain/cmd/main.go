package main

import (
	"fmt"
	"log"

	"github.com/miguelmota/go-ethereum-hdwallet"
)

func main() {
	// mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"
	mnemonic := "element urban soda endless beach celery scheme wet envelope east glory retire"
	fmt.Println("mnemonic:", mnemonic)

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal(err)
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("m/44'/60'/0'/0/0 Account address: %s\n", account.Address.Hex())

	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		log.Fatal(err)
	}
	privateKeyHex, err := wallet.PrivateKeyHex(account)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Private key m/44'/60'/0'/0/0 in hex: %s\n", privateKeyHex)

	publicKey, _ := wallet.PublicKeyHex(account)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Public key m/44'/60'/0'/0/0 in hex: %s\n", publicKey)

	_ = privateKey
	fmt.Println("=================")

	//第1号索引派生
	{
		path1 := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/1")
		account1, err := wallet.Derive(path1, true)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("m/44'/60'/0'/0/1 Account address: %s\n", account1.Address.Hex())

		privateKey1, err := wallet.PrivateKey(account1)
		if err != nil {
			log.Fatal(err)
		}
		privateKeyHex1, err := wallet.PrivateKeyHex(account1)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("m/44'/60'/0'/0/1 Private key in hex: %s\n", privateKeyHex1)

		publicKey1, _ := wallet.PublicKeyHex(account1)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("m/44'/60'/0'/0/1 Public key in hex: %s\n", publicKey1)

		_ = privateKey1
	}
}

/*

mnemonic: element urban soda endless beach celery scheme wet envelope east glory retire
m/44'/60'/0'/0/0 Account address: 0xe14D151e0511b61357DDe1B35a74E9c043c34C47
Private key m/44'/60'/0'/0/0 in hex: 4c88e6ccffec59b6c3df5ab51a4e6c42c421f58274d653d716aafd4aff376f5b
Public key m/44'/60'/0'/0/0 in hex: b97cf13c8758594fb59c14765f365d05b9e67539e8f50721f8f6b8401f13af93e623ee620d9de8058b4043a0bc8be99e9135b6aa1c10e9ca8e85e0c4828e3070
=================
m/44'/60'/0'/0/1 Account address: 0x4acea697f366C47757df8470e610a2d9B559DbBE
m/44'/60'/0'/0/1 Private key in hex: fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915
m/44'/60'/0'/0/1 Public key in hex: 553d2e5a5ad1ac9b2ae2dab3ddc28df74e1a549a753706715ec238e3e5c55008e45995b0d3271f8120890c74acc3602829207cefd432cfe1c1ca25767fd7a439

*/