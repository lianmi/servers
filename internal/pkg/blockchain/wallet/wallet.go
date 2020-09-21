package wallet

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/lianmi/servers/internal/pkg/blockchain/bip44"
	// "github.com/lianmi/servers/internal/pkg/blockchain/btcd/btcec"
	// "github.com/lianmi/servers/internal/pkg/blockchain/btcd/chaincfg"
	// "github.com/lianmi/servers/internal/pkg/blockchain/btcd/chaincfg/chainhash"
	// "github.com/lianmi/servers/internal/pkg/blockchain/btcd/txscript"
	// "github.com/lianmi/servers/internal/pkg/blockchain/btcd/wire"
	// "github.com/lianmi/servers/internal/pkg/blockchain/btcutil"
	"github.com/lianmi/servers/internal/pkg/blockchain/btcutil/hdkeychain"

	"errors"
	"github.com/lianmi/servers/internal/pkg/blockchain/bip39"
)

func CreateMnemonic(bitSize int) (string, error) {
	entropy, err := bip39.NewEntropy(bitSize)
	if err != nil {
		return "", err
	}

	return bip39.NewMnemonic(entropy)
}

func CreateSeed(mnemonic string, password string) []byte {
	return bip39.NewSeed(mnemonic, password)
}

func CreateMasterKeyFromSeed(seed []byte, net *chaincfg.Params) (*hdkeychain.ExtendedKey, error) {
	return hdkeychain.NewMaster(seed, net)
}

func CreateChangeKey(masterKey *hdkeychain.ExtendedKey, coinType uint32) (string, string, error) {
	changekey, err := bip44.NewChangeKeyFromMasterKey(masterKey, coinType, 0, 0)
	if err != nil {
		return "", "", err
	}

	// fmt.Println("changekey : ", changekey.String())
	changepubkey, _ := changekey.Neuter()
	// fmt.Println("changepubkey : ", changepubkey.String())

	return changekey.String(), changepubkey.String(), nil
}
