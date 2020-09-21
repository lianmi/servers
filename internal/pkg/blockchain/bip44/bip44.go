// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package bip44

import (
	"fmt"
	// "xiaomawallet/bip32"
	// "xiaomawallet/bip39"
	"github.com/lianmi/servers/internal/pkg/blockchain/btcutil/hdkeychain"
)

const Purpose uint32 = 0x8000002C //
const hardened uint32 = 0x80000000

//https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki
//https://github.com/satoshilabs/slips/blob/master/slip-0044.md
//https://github.com/FactomProject/FactomDocs/blob/master/wallet_info/wallet_test_vectors.md

const (
	TypeBitcoin               uint32 = 0x80000000
	TypeTestnet               uint32 = 0x80000001
	TypeLitecoin              uint32 = 0x80000002
	TypeDogecoin              uint32 = 0x80000003
	TypeReddcoin              uint32 = 0x80000004
	TypeDash                  uint32 = 0x80000005
	TypePeercoin              uint32 = 0x80000006
	TypeNamecoin              uint32 = 0x80000007
	TypeFeathercoin           uint32 = 0x80000008
	TypeCounterparty          uint32 = 0x80000009
	TypeBlackcoin             uint32 = 0x8000000a
	TypeNuShares              uint32 = 0x8000000b
	TypeNuBits                uint32 = 0x8000000c
	TypeMazacoin              uint32 = 0x8000000d
	TypeViacoin               uint32 = 0x8000000e
	TypeClearingHouse         uint32 = 0x8000000f
	TypeRubycoin              uint32 = 0x80000010
	TypeGroestlcoin           uint32 = 0x80000011
	TypeDigitalcoin           uint32 = 0x80000012
	TypeCannacoin             uint32 = 0x80000013
	TypeDigiByte              uint32 = 0x80000014
	TypeOpenAssets            uint32 = 0x80000015
	TypeMonacoin              uint32 = 0x80000016
	TypeClams                 uint32 = 0x80000017
	TypePrimecoin             uint32 = 0x80000018
	TypeNeoscoin              uint32 = 0x80000019
	TypeJumbucks              uint32 = 0x8000001a
	TypeziftrCOIN             uint32 = 0x8000001b
	TypeVertcoin              uint32 = 0x8000001c
	TypeNXT                   uint32 = 0x8000001d
	TypeBurst                 uint32 = 0x8000001e
	TypeMonetaryUnit          uint32 = 0x8000001f
	TypeZoom                  uint32 = 0x80000020
	TypeVpncoin               uint32 = 0x80000021
	TypeCanadaeCoin           uint32 = 0x80000022
	TypeShadowCash            uint32 = 0x80000023
	TypeParkByte              uint32 = 0x80000024
	TypePandacoin             uint32 = 0x80000025
	TypeStartCOIN             uint32 = 0x80000026
	TypeMOIN                  uint32 = 0x80000027
	TypeArgentum              uint32 = 0x8000002D
	TypeGlobalCurrencyReserve uint32 = 0x80000031
	TypeNovacoin              uint32 = 0x80000032
	TypeAsiacoin              uint32 = 0x80000033
	TypeBitcoindark           uint32 = 0x80000034
	TypeDopecoin              uint32 = 0x80000035
	TypeTemplecoin            uint32 = 0x80000036
	TypeAIB                   uint32 = 0x80000037
	TypeEDRCoin               uint32 = 0x80000038
	TypeSyscoin               uint32 = 0x80000039
	TypeSolarcoin             uint32 = 0x8000003a
	TypeSmileycoin            uint32 = 0x8000003b
	TypeEther                 uint32 = 0x8000003c //以太坊
	TypeEtherClassic          uint32 = 0x8000003d
	TypeOpenChain             uint32 = 0x80000040
	TypeOKCash                uint32 = 0x80000045
	TypeDogecoinDark          uint32 = 0x8000004d
	TypeElectronicGulden      uint32 = 0x8000004e
	TypeClubCoin              uint32 = 0x8000004f
	TypeRichCoin              uint32 = 0x80000050
	TypePotcoin               uint32 = 0x80000051
	TypeQuarkcoin             uint32 = 0x80000052
	TypeTerracoin             uint32 = 0x80000053
	TypeGridcoin              uint32 = 0x80000054
	TypeAuroracoin            uint32 = 0x80000055
	TypeIXCoin                uint32 = 0x80000056
	TypeGulden                uint32 = 0x80000057
	TypeBitBean               uint32 = 0x80000058
	TypeBata                  uint32 = 0x80000059
	TypeMyriadcoin            uint32 = 0x8000005a
	TypeBitSend               uint32 = 0x8000005b
	TypeUnobtanium            uint32 = 0x8000005c
	TypeMasterTrader          uint32 = 0x8000005d
	TypeGoldBlocks            uint32 = 0x8000005e
	TypeSaham                 uint32 = 0x8000005f
	TypeChronos               uint32 = 0x80000060
	TypeUbiquoin              uint32 = 0x80000061
	TypeEvotion               uint32 = 0x80000062
	TypeSaveTheOcean          uint32 = 0x80000063
	TypeBigUp                 uint32 = 0x80000064
	TypeGameCredits           uint32 = 0x80000065
	TypeDollarcoins           uint32 = 0x80000066
	TypeZayedcoin             uint32 = 0x80000067
	TypeDubaicoin             uint32 = 0x80000068
	TypeStratis               uint32 = 0x80000069
	TypeShilling              uint32 = 0x8000006a
	TypePiggyCoin             uint32 = 0x80000076
	TypeMonero                uint32 = 0x80000080
	TypeNavCoin               uint32 = 0x80000082
	TypeFactomFactoids        uint32 = 0x80000083
	TypeFactomEntryCredits    uint32 = 0x80000084
	TypeZcash                 uint32 = 0x80000085
	TypeLisk                  uint32 = 0x80000086
)

/*
func NewAddressKeyFromMnemonic(mnemonic string, coin, account, chain, address uint32) (*hdkeychain.ExtendedKey, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, err
	}

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}

	fmt.Println("masterKey :", masterKey.String())
	return NewAddressKeyFromMasterKey(masterKey, coin, account, chain, address)
}
*/
func NewChangeKeyFromMasterKey(masterKey *hdkeychain.ExtendedKey, coin, k, chain uint32) (*hdkeychain.ExtendedKey, error) {

	purpose, err := masterKey.Child(Purpose)
	if err != nil {
		return nil, err
	}
	fmt.Println("m/44' :", purpose.String())

	// hardenedStr := fmt.Sprint(hardened)
	// fmt.Println("hardened=", hardenedStr)

	//此句有问题，hardened=0x80000000， coin由外部传入，值是:
	// coin_type, err := purpose.Child(hardened + coin)
	coin_type, err := purpose.Child(hardened + coin)
	if err != nil {
		return nil, err
	}

	fmt.Println("m/44'/cointype :", coin_type.String())
	account, err := coin_type.Child(hardened + k)
	if err != nil {
		return nil, err
	}
	fmt.Println("m/44'/cointype/0' :", account.String())
	change, err := account.Child(chain)
	if err != nil {
		return nil, err
	}
	fmt.Println("m/44'/cointype/0'/0 :", change.String())

	// //索引=0
	// addresskey, err := change.Child(0)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("m/44'/1‘/0'/0/0 :", addresskey.String())

	return change, nil
}

func NewAddressKeyFromMasterKey(masterKey *hdkeychain.ExtendedKey, coin, k, chain, address uint32) (*hdkeychain.ExtendedKey, error) {

	purpose, err := masterKey.Child(Purpose)
	if err != nil {
		return nil, err
	}
	fmt.Println("m/44' :", purpose.String())
	//此句有问题，不应该传hardened + coin， 只用coid就足够了
	// coin_type, err := purpose.Child(hardened + coin)
	coin_type, err := purpose.Child(coin)
	if err != nil {
		return nil, err
	}
	fmt.Println("m/44'/cointype :", coin_type.String())
	account, err := coin_type.Child(hardened + k)
	if err != nil {
		return nil, err
	}
	fmt.Println("m/44'/cointype/0' :", account.String())
	chainkey, err := account.Child(chain)
	if err != nil {
		return nil, err
	}
	fmt.Println("m/44'/cointype/0'/0 :", chainkey.String())
	addresskey, err := chainkey.Child(address)
	if err != nil {
		return nil, err
	}

	fmt.Println("m/44'/cointype/0'/0/0 :", addresskey.String())
	return addresskey, nil
}
