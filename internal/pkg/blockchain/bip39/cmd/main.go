package main

/*

mnemonic:  job gravity goose boring filter lyrics source giant throw dismiss film emotion margin depend ostrich peanut exist version unfold logic cause protect section drama

account Address:  0xC50Fe56057B5D6Ab4b714C54d72C8e3018975D5D
Private key0 of account in hex: 387153a31bf48456fed325e1a5be9e17c1c87e00cd5bac8721db3b0cc79a1d74
Public key0 of account  in hex: 906abda2050da89224a1d9e13d64f38b14de1f0f46b2043354f7032d2ba1ebdb0b7a88bb40700ce2a0deca6e9e28524f2bff3f63654dc6e94561ed5babedf5eb

account1 Address:  0x1826654168d449004794C1d6F092d5E3F0F8365A
Private key1 of account in hex: 387153a31bf48456fed325e1a5be9e17c1c87e00cd5bac8721db3b0cc79a1d74
Public key1 of account  in hex: 906abda2050da89224a1d9e13d64f38b14de1f0f46b2043354f7032d2ba1ebdb0b7a88bb40700ce2a0deca6e9e28524f2bff3f63654dc6e94561ed5babedf5eb

*/
import (
	"fmt"
	"github.com/lianmi/servers/internal/pkg/blockchain/bip39"
	"log"
	// "math/big"
	// "os"

	// "github.com/davecgh/go-spew/spew"
	// "github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/core/types"
	"github.com/miguelmota/go-ethereum-hdwallet"
)

// bitSize has to be a multiple 32 and be within the inclusive range of {128, 256}
func CreateMnemonic(bitSize int) (string, error) {
	entropy, err := bip39.NewEntropy(bitSize)
	if err != nil {
		return "", err
	}

	return bip39.NewMnemonic(entropy)
}

func main() {
	/*
		mnemonic, err := CreateMnemonic(256)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	*/
	mnemonic := "element urban soda endless beach celery scheme wet envelope east glory retire"
	fmt.Println("mnemonic: ", mnemonic)

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal(err)
	}

	// 叶子路径
	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("account Address: ", account.Address.Hex())
	privateKey, err := wallet.PrivateKeyHex(account)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Private key0 of account in hex: %s\n", privateKey)

	publicKey, _ := wallet.PublicKeyHex(account)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Public key0 of account  in hex: %s\n", publicKey)

	/*
		path1 := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/1")
		account1, err := wallet.Derive(path1, true)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("account1 Address: ", account1.Address.String())

		privateKey1, err := wallet.PrivateKeyHex(account)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Private key1 of account in hex: %s\n", privateKey1)

		publicKey1, _ := wallet.PublicKeyHex(account)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Public key1 of account  in hex: %s\n", publicKey1)
	*/
	/*
		nonce := uint64(0)
		value := big.NewInt(1000000000000000000)
		toAddress := common.HexToAddress("0x0")
		gasLimit := uint64(21000)
		gasPrice := big.NewInt(21000000000)
		var data []byte

		tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
		signedTx, err := wallet.SignTx(account, tx, nil)
		if err != nil {
			log.Fatal(err)
		}

		// a deep pretty printer
		spew.Dump(signedTx)
	*/
}
