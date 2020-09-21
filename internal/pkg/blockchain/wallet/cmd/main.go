package main
import (
	"github.com/lianmi/servers/internal/pkg/blockchain/bip39"
	"github.com/lianmi/servers/internal/pkg/blockchain/bip44"
	"github.com/lianmi/servers/internal/pkg/blockchain/wallet"
)

//热根 助记符 :  element urban soda endless beach celery scheme wet envelope east glory retire
//服务端 助记符 :  someone author recipe spider ready exile occur volume relax song inner inform
//冷根 助记符 :  cloth have cage erase shrug slot album village surprise fence erode direct
func CreateRootPairKeys() (hotPair PairKeys, serverPair PairKeys, coolPair PairKeys, err error) {

	seed1 := wallet.CreateSeed(mnemonic_hot, "unitwallet")
	seed1Hex := hex.EncodeToString(seed1)
	fmt.Println("seed1Hex=", seed1Hex)

	masterkey1, err := wallet.CreateMasterKeyFromSeed(seed1, MAINNET_ID)
	if err != nil {
		fmt.Println("Client  server CreateMasterKeyFromSend error : ", err)
	}
	strprv1 := masterkey1.String()
	pub1, _ := masterkey1.Neuter()
	strpub1 := pub1.String()
	fmt.Println("热根主私钥 MasterKeyprv1 : ", strprv1)
	fmt.Println("热根主公钥 MasterKeypub1 : ", strpub1)
	// tesstrprv1s = append(tesstrprv1s,strprv1)
	// teststrpub1s = append(teststrpub1s,strpub1)

	//倒数第二层 m/44'/0'/0'/0
	changeprv1, changepub1, err := wallet.CreateChangeKey(masterkey1, bip44.TypeBitcoin)
	fmt.Println("热根 changeprv1 : ", changeprv1)
	fmt.Println("热根 changepub1 : ", changepub1)

	seed3 := wallet.CreateSeed(mnemonic_cool, "unitwallet")
	seed3Hex := hex.EncodeToString(seed3)
	fmt.Println("seed3Hex=", seed3Hex)

	masterkey3, err := wallet.CreateMasterKeyFromSeed(seed3, MAINNET_ID)
	if err != nil {
		fmt.Println("Client  server CreateMasterKeyFromSend error : ", err)
	}
	strprv3 := masterkey3.String()
	pub3, _ := masterkey3.Neuter()
	strpub3 := pub3.String()
	fmt.Println("冷根主私钥 MasterKeyprv3 : ", strprv3)
	fmt.Println("冷根主公钥 MasterKeypub3 : ", strpub3)

	//bip44.TypeBitcoin 主网，
	changeprv3, changepub3, err := wallet.CreateChangeKey(masterkey3, bip44.TypeBitcoin)
	fmt.Println("冷根 changeprv3 : ", changeprv3)
	fmt.Println("冷根 changepub3 : ", changepub3)

	seed2 := wallet.CreateSeed(mnemonic_server, "unitwallet")
	seed2Hex := hex.EncodeToString(seed2)
	fmt.Println("seed2Hex=", seed2Hex)

	masterkey2, err := wallet.CreateMasterKeyFromSeed(seed2, MAINNET_ID)
	if err != nil {
		fmt.Println("Client  server CreateMasterKeyFromSend error : ", err)
	}
	strprv2 := masterkey2.String()
	pub2, _ := masterkey2.Neuter()
	strpub2 := pub2.String()
	fmt.Println("服务端主私钥 MasterKeyprv2 : ", strprv2)
	fmt.Println("服务端主私钥 MasterKeypub2 : ", strpub2)

	changeprv2, changepub2, err := wallet.CreateChangeKey(masterkey2, bip44.TypeBitcoin)
	fmt.Println("服务端 changeprv2 : ", changeprv2)
	fmt.Println("服务端 changepub2 : ", changepub2)

	hotPair = PairKeys{
		PrivKey: changeprv1,
		PubKey:  changepub1,
	}
	serverPair = PairKeys{
		PrivKey: changeprv2,
		PubKey:  changepub2,
	}
	coolPair = PairKeys{
		PrivKey: changeprv3,
		PubKey:  changepub3,
	}

	return hotPair, serverPair, coolPair, nil
}

func main() {
	hotPair, serverPair, coolPair, err := CreateRootPairKeys()
	if err != nil {
		log.Println("CreateRootPairKeys error", CreateRootPairKeys)
		return
	}
	log.Println("############")
	log.Println("BIP44热根子私钥(m/44'/0'/0'/0)：", hotPair.PrivKey)
	log.Println("BIP44热根子公钥(m/44'/0'/0'/0)：", hotPair.PubKey)

	_ = serverPair
	_ = coolPair

	// log.Println("############")
	// log.Println("BIP44服务端子私钥(m/44'/1'/0'/0)：", serverPair.PrivKey)
	// log.Println("BIP44服务端子公钥(m/44'/1'/0'/0)：", serverPair.PubKey)

	// log.Println("############")
	// log.Println("BIP44冷根子私钥(m/44'/1'/0'/0)：", coolPair.PrivKey)
	// log.Println("BIP44冷根子公钥(m/44'/1'/0'/0)：", coolPair.PubKey)

	privKey1, err := wallet.CreateBTCAddressFromPriv(hotPair.PrivKey, uint32(0))
	if err != nil {
		fmt.Errorf("#1 failed to CreateBTCAddressFromPriv:  %v",
			err)
		os.Exit(1)
	}
	privKey1_wif, err := btcutil.NewWIF(privKey1, MAINNET_ID, true)
	if err != nil {
		fmt.Errorf("#1 failed to NewWIF:  %v",
			err)
		os.Exit(1)
	}

	pk1, btcaddress1, err := wallet.CreateBTCAddressFromPub(hotPair.PubKey, uint32(0), MAINNET_ID)
	if err != nil {
		fmt.Errorf("#1 failed to CreateBTCAddressFromPub:  %v",
			err)
		os.Exit(1)
	}
	addressPubKey1, err := btcutil.NewAddressPubKey(pk1.SerializeCompressed(), MAINNET_ID)
	if err != nil {
		fmt.Errorf("failed to make address 3 for  %v",
			err)
	}

	fmt.Println("热根子私钥 privKey1 hex: ", hex.EncodeToString(privKey1.Serialize()))
	fmt.Println("热根子私钥 privKey1 wif : ", privKey1_wif)
	fmt.Println("热根子公钥 addressPubKey1 : ", addressPubKey1.String())
	fmt.Println("热根子地址 btcaddress1 : ", btcaddress1)

	/*
		privKey2, err := wallet.CreateBTCAddressFromPriv(serverPair.PrivKey, uint32(0))
		if err != nil {
			fmt.Errorf("#2 failed to CreateBTCAddressFromPriv:  %v",
				err)
			os.Exit(1)
		}
		privKey2_wif, err := btcutil.NewWIF(privKey2, MAINNET_ID, true)
		if err != nil {
			fmt.Errorf("#1 failed to NewWIF:  %v",
				err)
			os.Exit(1)
		}

		privKey3, err := wallet.CreateBTCAddressFromPriv(coolPair.PrivKey, uint32(0))
		if err != nil {
			fmt.Errorf("#3 failed to CreateBTCAddressFromPriv:  %v",
				err)
			os.Exit(1)
		}
		privKey3_wif, err := btcutil.NewWIF(privKey3, MAINNET_ID, true)
		if err != nil {
			fmt.Errorf("#1 failed to NewWIF:  %v",
				err)
			os.Exit(1)
		}

		pk2, btcaddress2, err := wallet.CreateBTCAddressFromPub(serverPair.PubKey, uint32(0), MAINNET_ID)
		if err != nil {
			fmt.Errorf("#2 failed to CreateBTCAddressFromPub:  %v",
				err)
			os.Exit(1)
		}
		pk3, btcaddress3, err := wallet.CreateBTCAddressFromPub(coolPair.PubKey, uint32(0), MAINNET_ID)
		if err != nil {
			fmt.Errorf("#3 failed to CreateBTCAddressFromPub:  %v",
				err)
			os.Exit(1)
		}


		address2, err := btcutil.NewAddressPubKey(pk2.SerializeCompressed(), MAINNET_ID)
		if err != nil {
			fmt.Errorf("failed to make address 3 for  %v",
				err)
		}

		address3, err := btcutil.NewAddressPubKey(pk3.SerializeCompressed(), MAINNET_ID)
		if err != nil {
			fmt.Errorf("failed to make address 3 for  %v",
				err)
		}

		fmt.Println("热根子私钥 privKey1 hex: ", hex.EncodeToString(privKey1.Serialize()))
		fmt.Println("热根子私钥 privKey1 wif : ", privKey1_wif)
		fmt.Println("热根子公钥 addressPubKey1 长度: ", address1.GetLength())
		fmt.Println("热根子公钥 addressPubKey1 : ", address1.String())

		fmt.Println("冷根子私钥 privKey3 hex: ", hex.EncodeToString(privKey3.Serialize()))
		fmt.Println("冷根子私钥 privKey3 wif : ", privKey3_wif)
		fmt.Println("冷根子公钥 addressPubKey3 长度: ", address3.GetLength())
		fmt.Println("冷根子公钥 addressPubKey3 : ", address3.String())

		fmt.Println("服务端子私钥 privKey2 hex: ", hex.EncodeToString(privKey2.Serialize()))
		fmt.Println("服务端子私钥 privKey2 wif : ", privKey2_wif)
		fmt.Println("服务端子公钥 addressPubKey2 长度: ", address2.GetLength())
		fmt.Println("服务端子公钥 addressPubKey2 : ", address2.String())

		fmt.Println("热根地址 address1 : ", btcaddress1)
		fmt.Println("服务端地址 address2 : ", btcaddress2)
		fmt.Println("冷根 address3 : ", btcaddress3)

		fmt.Println("********************************** 生成多签地址 *************************************** ")
		pubkeys := []*btcutil.AddressPubKey{address1, address3, address2}
		MultiSigAddress, redeemScript, err := wallet.CreateMultiSigAddress(pubkeys, 2, MAINNET_ID)

		fmt.Println("address_index : ", 0)
		fmt.Println("MultiSigAddress : ", MultiSigAddress)
		fmt.Println("redeemScript : ", redeemScript)
		fmt.Println("        ")

		//创建交易
		GenerateUTXO(privKey1, privKey2)
		_ = privKey3
	*/
}

