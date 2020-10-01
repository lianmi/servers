package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"strings"
	// /Users/mac/developments/lianmi/lm-cloud/servers/internal/pkg/blockchain/lnmc/contracts/ERC20
	ERC20 "github.com/lianmi/servers/internal/pkg/blockchain/lnmc/contracts/ERC20"
	MultiSig "github.com/lianmi/servers/internal/pkg/blockchain/lnmc/contracts/MultiSig"
)

// 0xa7cc1ae7199cce8aa1354059953f6559cf57869f 的key文件，发币地址
const key = "UTC--2020-09-19T15-07-00.413330000Z--a7cc1ae7199cce8aa1354059953f6559cf57869f"

//部署多签合约
func deployMultiSig() {
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	// 合约部署
	data, _ := ioutil.ReadFile(key)
	auth, err := bind.NewTransactor(strings.NewReader(string(data)), "123456")
	if err != nil {
		log.Fatalf("Failed to create authorized transactor:%v \n", err)
	}

	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(gasPrice.String())
	gasPrice.Mul(gasPrice, big.NewInt(10))
	fmt.Println(gasPrice.String())
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(3000000) // in units 注意，不能太少
	auth.GasPrice = gasPrice

	//A的私钥
	privateKeyA, err := crypto.HexToECDSA("91e5f2d81444905af5f94d6b36be36d69363420b9edd59808caec17830d50ff1")
	if err != nil {
		log.Fatal(err)
	}
	publicKeyA := privateKeyA.Public()
	publicKeyECDSAA, ok := publicKeyA.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddressA := crypto.PubkeyToAddress(*publicKeyECDSAA)

	//B的私钥
	privateKeyB, err := crypto.HexToECDSA("b65e1f6e3b449c35c18518cfdf8de3c361ccf6f4a51817e0709a917fac688423")
	if err != nil {
		log.Fatal(err)
	}
	publicKeyB := privateKeyB.Public()
	publicKeyECDSAB, ok := publicKeyB.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddressB := crypto.PubkeyToAddress(*publicKeyECDSAB)
	fmt.Println(fromAddressA.String(), fromAddressB.String())
	//return
	address, tx, _, err := MultiSig.DeployMultiSig(
		auth,
		client,
		fromAddressA, //A 账号地址
		fromAddressB, //B 账号地址
		common.HexToAddress("0xdeb284d75f757ce5e3c5de349732c05baa53584f"), //ERC20发币地址
	)
	if err != nil {
		log.Fatalf("deploy %v \n", err)
	}
	fmt.Println("Contract pending deploy: ", address.String(), tx.Hash().String())

	//TODO 监听，直到合约部署成功,如果失败，则提示

}

//将一定数量amount的代币转账到多签合约账户
func sendTokenToMultisigContractAddress(source, target string, amount int64) {
	blockchain, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	//使用合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(source), blockchain)
	if err != nil {
		log.Fatalf("conn contract: %v \n", err)
	}

	data, _ := ioutil.ReadFile(key)
	auth, err := bind.NewTransactor(strings.NewReader(string(data)), "123456")
	if err != nil {
		log.Fatalf("Failed to create authorized transactor:%v \n", err)
	}

	//调用合约里的转账函数
	tx, err := contract.Transfer(&bind.TransactOpts{
		From:   auth.From,
		Signer: auth.Signer,
		Value:  nil,
	}, common.HexToAddress(target), big.NewInt(amount))
	if err != nil {
		log.Fatalf("TransferFrom err: %v \n", err)
	}
	fmt.Printf("tx sent: %s \n", tx.Hash().Hex())

	//TODO 监听，直到转账成功,如果失败，则提示
}

func main01() {
	//部署 多签合约
	// deployMultiSig()
	/*
	   1000000000
	   10000000000
	   0x6d9CFbC20E1b210d25b84F83Ba546ea4264DA538 0xac243c2FED19d085bF682d0D74e677c1d9911e83
	   Contract pending deploy:  0x516E44Bb27B8bfA55A53b6b79EF3bfc265aC3057 0x69e3eae90d8e94f1aeb07b255f0e619f8818660eeedf4f452bba3f95060cd068

	*/
	// sendTokenToMultisigContractAddress("0xdeb284d75f757ce5e3c5de349732c05baa53584f", "0x516E44Bb27B8bfA55A53b6b79EF3bfc265aC3057", 200)
	//tx sent: 0x2361c8d8a3374d63b24904b6ce3a13d95f38e0cf379bc5ba99e552f0c53566fb

	//查询余额
	// querySendAndReceive("0xdeb284d75f757ce5e3c5de349732c05baa53584f", "0x516E44Bb27B8bfA55A53b6b79EF3bfc265aC3057")

	//给A转账 代币200
	// sendTokenToMultisigContractAddress("0xdeb284d75f757ce5e3c5de349732c05baa53584f", "0x6d9CFbC20E1b210d25b84F83Ba546ea4264DA538", 200)
	//tx sent:

	//查询A余额
	// querySendAndReceive("0xdeb284d75f757ce5e3c5de349732c05baa53584f", "0x6d9CFbC20E1b210d25b84F83Ba546ea4264DA538")

	//A调用智能合约进行转账
	// transfer(
	// 	"0x516E44Bb27B8bfA55A53b6b79EF3bfc265aC3057",                       //第一步部署的多签合约地址
	// 	"91e5f2d81444905af5f94d6b36be36d69363420b9edd59808caec17830d50ff1", //A私钥
	// 	"0xba8d69ba4d65802039cfe2ae373072639026d457",                       //C账号地址
	// 	50)

	/*
	 输出：

	*/

	//B审核
	// transfer(
	// 	"0x516E44Bb27B8bfA55A53b6b79EF3bfc265aC3057",                       //第一步部署的多签合约地址
	// 	"b65e1f6e3b449c35c18518cfdf8de3c361ccf6f4a51817e0709a917fac688423", //B私钥
	// 	"0xba8d69ba4d65802039cfe2ae373072639026d457",                       //C账号地址
	// 	50)

	//查询 C 账号的ERC20
	// querySendAndReceive("0xdeb284d75f757ce5e3c5de349732c05baa53584f", "0xba8d69ba4d65802039cfe2ae373072639026d457")

	//查询多签合约里剩余的代币
	// querySendAndReceive("0xdeb284d75f757ce5e3c5de349732c05baa53584f", "0x516E44Bb27B8bfA55A53b6b79EF3bfc265aC3057")

	//查询C账号的ERC20
	// getBalance("0xdeb284d75f757ce5e3c5de349732c05baa53584f", "0xba8d69ba4d65802039cfe2ae373072639026d457")

}
