package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	// "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	ERC20 "github.com/lianmi/servers/internal/pkg/blockchain/lnmc/contracts/ERC20"
	MultiSig "github.com/lianmi/servers/internal/pkg/blockchain/lnmc/contracts/MultiSig"
	"github.com/lianmi/servers/internal/pkg/blockchain/util"
)

const (
	WSURIIPC  = "ws://127.0.0.1:8546"
	RedisAddr = "127.0.0.1:6379"

	PASSWORD = "LianmiSky8900388"
	GASLIMIT = 5000000 //6000000

	PrivateKeyAHEX = "91e5f2d81444905af5f94d6b36be36d69363420b9edd59808caec17830d50ff1" //用户A私钥
	AddressAHEX    = "0x6d9CFbC20E1b210d25b84F83Ba546ea4264DA538"                       //用户A地址

	PrivateKeyBHEX = "b65e1f6e3b449c35c18518cfdf8de3c361ccf6f4a51817e0709a917fac688423" //用户B私钥
	AddressBHEX    = "0xac243c2FED19d085bF682d0D74e677c1d9911e83"                       //用户B地址

	AddressCHEX = "0xBa8d69ba4D65802039cfE2ae373072639026D457" //用户C地址
	AddressDHEX = "0x59aC768b416C035c8DB50B4F54faaa1E423c070D" //用户D地址
)

/*
发币合约部署
第1号叶子:
privateKey: fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915
address: 0x4acea697f366C47757df8470e610a2d9B559DbBE
*/

func deploy(privateKeyHex string) error {
	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	redisConn, err := redis.Dial("tcp", RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()
	//
	erc20DeployContractAddress, _ := redis.String(redisConn.Do("GET", "ERC20DeployContractAddress"))
	if err != nil {
		log.Fatal(err)
		return err
	}
	if erc20DeployContractAddress != "" {
		return errors.New("erc20DeployContractAddress had deployed!")
	}
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)       // in wei
	auth.GasLimit = uint64(GASLIMIT) // in units
	auth.GasPrice = gasPrice

	address, deployTx, _, err := ERC20.DeployERC20Token(
		auth,
		client,
		big.NewInt(1000000000000), //10000亿枚，一枚等于1分钱
		"LianmiCoin",
		"LNMC",
	)
	if err != nil {
		log.Fatalf("deploy %v \n", err)
	}
	fmt.Printf("Contract pending deploy:  0x%x \n", address.String())

	done := WaitForBlockCompletation(client, deployTx.Hash().Hex())
	if done >= 1 {
		log.Println("交易完成")
		tx, isPending, err := client.TransactionByHash(context.Background(), deployTx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		log.Println("发币合约部署成功, 交易哈希: ", tx.Hash().Hex()) //
		log.Println("isPending: ", isPending)            // false
		//保存到redis里
		_, err = redisConn.Do("SET", "ERC20DeployContractAddress", address.String())
		if err != nil {
			log.Fatal(err)
		}

	} else {
		log.Println("发币合约部署失败")
	}

	// headers := make(chan *types.Header)
	// sub, err := client.SubscribeNewHead(context.Background(), headers)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// for {
	// 	select {
	// 	case err := <-sub.Err():
	// 		log.Fatal(err)
	// 	case header := <-headers:
	// 		fmt.Println(header.Hash().Hex()) // 0xf918192c4dd05834ebdee15920225f97d4ed350c829b007ae8ca95217f282f3e

	// 		block, err := client.BlockByHash(context.Background(), header.Hash())
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}

	// 		fmt.Println(block.Number().Uint64())   // 1886
	// 		fmt.Println(block.Time())              // 1601576368
	// 		fmt.Println(block.Nonce())             // 3489525387448087149
	// 		fmt.Println(len(block.Transactions())) // 1
	// 	}
	// }

	return nil
}

//传参：账户地址
func getTokenBalance(accountAddress string) error {
	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	redisConn, err := redis.Dial("tcp", RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()
	//
	erc20DeployContractAddress, _ := redis.String(redisConn.Do("GET", "ERC20DeployContractAddress"))
	if err != nil {
		log.Fatal(err)
		return err
	}
	if erc20DeployContractAddress == "" {
		return errors.New("erc20DeployContractAddress is required")
	}

	//使用合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(erc20DeployContractAddress), client)
	if err != nil {
		log.Fatalf("conn contract: %v \n", err)
	}

	//余额查询
	accountBalance, err := contract.BalanceOf(nil, common.HexToAddress(accountAddress))
	if err != nil {
		log.Fatalf("get Balances err: %v \n", err)
	}
	fmt.Println("Token of LNMC:", accountBalance)
	return nil

}

//ERC20代币余额查询， 传参1是发送者合约地址，传参2是接收者账号地址
func querySendAndReceive(sender, receiver string) {

	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	//使用合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(sender), client)
	if err != nil {
		log.Fatalf("conn contract: %v \n", err)
	}
	// data, _ := ioutil.ReadFile(key)
	// auth, err := bind.NewTransactor(strings.NewReader(string(data)), "123456")
	// if err != nil {
	// 	log.Fatalf("Failed to create authorized transactor:%v \n", err)
	// }

	// var accountBalance = big.NewInt(0)
	// if accountBalance, err = contract.BalanceOf(nil, auth.From); err != nil {
	// 	log.Fatalf("get Balances err: %v \n", err)
	// }
	// fmt.Println("发送方余额: ", accountBalance)

	// target是接收地址
	if accountBalance, err := contract.BalanceOf(nil, common.HexToAddress(receiver)); err != nil {
		log.Fatalf("get Balances err: %v \n", err)
	} else {

		fmt.Println("接收方余额: ", accountBalance)
	}

}

func queryTx(txHex string) {
	txHash := common.HexToHash(txHex)
	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(tx.Hash().Hex()) // 0x40aa1ed6e2af939a9cc2f711a51cea0f21bdba3f146530f270956dbe3b454dd8
	fmt.Println(isPending)       // false
}

//查询Eth余额
func queryETH(accountHex string) {

	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	account := common.HexToAddress(accountHex)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(accountHex, ":", balance) // 1000000000000000000

}

//裸交易结构体
type RawDesc struct {
	//nonce
	Nonce uint64 `protobuf:"fixed64,1,opt,name=nonce,proto3" json:"nonce,omitempty"`
	// gas价格
	GasPrice uint64 `protobuf:"fixed64,2,opt,name=gasPrice,proto3" json:"gasPrice,omitempty"`
	// 最低gas
	GasLimit uint64 `protobuf:"fixed64,3,opt,name=gasLimit,proto3" json:"gasLimit,omitempty"`
	//链id
	ChainID uint64 `protobuf:"fixed64,4,opt,name=chainID,proto3" json:"chainID,omitempty"`
	// 交易数据
	Txdata []byte `protobuf:"bytes,5,opt,name=txdata,proto3" json:"txdata,omitempty"`
	//多签合约地址
	ContractAddress string `protobuf:"bytes,6,opt,name=contractAddress,proto3" json:"contractAddress,omitempty"`
	//ether，设为0
	Value uint64 `protobuf:"fixed64,7,opt,name=value,proto3" json:"value,omitempty"`
}

//检查交易打包状态
func checkTransactionReceipt(_txHash string) int {
	// client, err := ethclient.Dial("http://127.0.0.1:8545")
	client, err := ethclient.Dial(WSURIIPC)
	txHash := common.HexToHash(_txHash)
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return (-1)
	}
	log.Println(receipt.Logs)
	return (int(receipt.Status))
}

//订阅并检测交易是否成功
func WaitForBlockCompletation(wsClient *ethclient.Client, hashToRead string) int {
	headers := make(chan *types.Header)
	sub, err := wsClient.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Println("Err: ", err)
			_ = err
			return -1
		case header := <-headers:
			log.Println(header.TxHash.Hex())
			transactionStatus := checkTransactionReceipt(hashToRead)
			if transactionStatus == 0 {
				log.Println("transactionStatus == 0")
				//FAILURE
				sub.Unsubscribe()
				return 0
			} else if transactionStatus == 1 {
				//SUCCESS
				block, err := wsClient.BlockByHash(context.Background(), header.Hash())
				if err != nil {
					log.Fatal(err)
				}
				log.Println("区块: ", block.Hash().Hex())
				log.Println("区块编号: ", block.Number().Uint64())
				queryTransactionByBlockNumber(block.Number().Uint64())
				sub.Unsubscribe()
				return 1
			}
		}
	}
}

//查询交易
func queryTransactionByBlockNumber(number uint64) {
	// client, err := ethclient.Dial("http://127.0.0.1:8545")
	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatal(err)
	}
	blockNumber := big.NewInt(int64(number))
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("=========queryTransactionByBlockNumber start==========")
	for _, tx := range block.Transactions() {
		log.Println("tx.Hash: ", tx.Hash().Hex())            //
		log.Println("tx.Value: ", tx.Value().String())       // 10000000000000000
		log.Println("tx.Gas: ", tx.Gas())                    // 6000000
		log.Println("tx.GasPrice: ", tx.GasPrice().Uint64()) // 1000000000

		// cost := tx.Gas() * tx.GasPrice().Uint64() //计算交易所需要支付的总费用
		gasCost := util.CalcGasCost(tx.Gas(), tx.GasPrice()) //计算交易所需要支付的总费用
		log.Println("交易总费用(Wei): ", gasCost)                 //6000000000000000Wei
		ethAmount := util.ToDecimal(gasCost, 18)
		log.Println("交易总费用(eth): ", ethAmount) //0.003Eth
		log.Println("tx.Nonce: ", tx.Nonce())
		log.Println("tx.Data: ", tx.Data())
		log.Println("tx.Data Hex: ", hex.EncodeToString(tx.Data()))
		// log.Println("tx.To: ", tx.To().Hex()) //目标地址

		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		if msg, err := tx.AsMessage(types.NewEIP155Signer(chainID)); err == nil {
			log.Println("tx.From: ", msg.From().Hex()) // 总账号
		}

		receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		log.Println(receipt.Status) // 1
	}
	log.Println("=========queryTransactionByBlockNumber end==========")
}

//部署多签合约
func deployMultiSig(privateKeyHex string) error {

	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
		return err
	}

	redisConn, err := redis.Dial("tcp", RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()
	//
	erc20DeployContractAddress, _ := redis.String(redisConn.Do("GET", "ERC20DeployContractAddress"))
	if err != nil {
		log.Fatal(err)
		return err
	}
	if erc20DeployContractAddress == "" {
		return errors.New("erc20DeployContractAddress is required")
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal(err)
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return err
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
		return err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
		return err
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)       // in wei
	auth.GasLimit = uint64(GASLIMIT) // in units
	auth.GasPrice = gasPrice

	//用户A的私钥
	privateKeyA, err := crypto.HexToECDSA(PrivateKeyAHEX)
	if err != nil {
		log.Fatal(err)
		return err
	}
	publicKeyA := privateKeyA.Public()
	publicKeyECDSAA, ok := publicKeyA.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return err
	}
	fromAddressA := crypto.PubkeyToAddress(*publicKeyECDSAA)

	// 商户B的私钥
	privateKeyB, err := crypto.HexToECDSA(PrivateKeyBHEX)
	if err != nil {
		log.Fatal(err)
		return err
	}
	publicKeyB := privateKeyB.Public()
	publicKeyECDSAB, ok := publicKeyB.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return err
	}
	fromAddressB := crypto.PubkeyToAddress(*publicKeyECDSAB)
	fmt.Println("fromAddressA: ", fromAddressA.String(), "fromAddressB: ", fromAddressB.String())

	address, deployMultiSigTx, _, err := MultiSig.DeployMultiSig(
		auth,
		client,
		fromAddressA, //A 账号地址
		fromAddressB, //B 账号地址
		common.HexToAddress(erc20DeployContractAddress), //ERC20发币地址
	)
	if err != nil {
		log.Fatalf("deploy %v \n", err)
		return err
	}
	fmt.Printf("Contract pending deploy: %s, SigTx Hash: %s\n", address.String(), deployMultiSigTx.Hash().String())

	//TODO 监听，直到合约部署成功,如果失败，则提示

	done := WaitForBlockCompletation(client, deployMultiSigTx.Hash().Hex())
	if done == 1 {
		log.Println("交易完成")
		tx, isPending, err := client.TransactionByHash(context.Background(), deployMultiSigTx.Hash())
		if err != nil {
			log.Fatal(err)
			return err
		}

		log.Println("多签合约部署成功, 交易哈希: ", tx.Hash().Hex()) //
		log.Println("isPending: ", isPending)            // false

	} else {
		log.Println("多签合约部署失败")

	}
	return nil
}

//将一定数量amount的代币由 总发币合约地址 转账到多签合约账户, soure是用户A
func sendTokenToMultisigContractAddress(sourcePrivateKey, target string, amount string) error {
	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
		return err
	}

	redisConn, err := redis.Dial("tcp", RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()
	//
	erc20DeployContractAddress, _ := redis.String(redisConn.Do("GET", "ERC20DeployContractAddress"))
	if err != nil {
		log.Fatal(err)
		return err
	}
	if erc20DeployContractAddress == "" {
		return errors.New("erc20DeployContractAddress is required")
	}

	//使用总发币的合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(erc20DeployContractAddress), client)
	if err != nil {
		log.Fatalf("conn contract: %v \n", err)
		return err
	}

	//A的私钥
	privateKey, err := crypto.HexToECDSA(sourcePrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	auth := bind.NewKeyedTransactor(privateKey)

	value := new(big.Int)
	value.SetString(amount, 10)

	//调用合约里的转账函数
	transferTx, err := contract.Transfer(&bind.TransactOpts{
		From:   auth.From,
		Signer: auth.Signer,
		Value:  nil,
	}, common.HexToAddress(target), value)
	if err != nil {
		log.Fatalf("TransferFrom err: %v \n", err)
	}
	log.Printf("tx sent: %s \n", transferTx.Hash().Hex())

	//TODO 监听，直到转账成功,如果失败，则提示
	done := WaitForBlockCompletation(client, transferTx.Hash().Hex())
	if done == 1 {
		tx, isPending, err := client.TransactionByHash(context.Background(), transferTx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		log.Println("代币从A转账到多签合约账户成功, 交易哈希: ", tx.Hash().Hex()) //
		log.Println("isPending: ", isPending)                   // false

	} else {
		log.Println("代币转账到多签合约账户失败")
	}
	return nil
}

//ERC20代币余额查询， 传参1是发送者合约地址，传参2是接收者账号地址
func querySendAndReceive2(sender, receiver string) error {

	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	redisConn, err := redis.Dial("tcp", RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()
	//
	erc20DeployContractAddress, _ := redis.String(redisConn.Do("GET", "ERC20DeployContractAddress"))
	if err != nil {
		log.Fatal(err)
		return err
	}
	if erc20DeployContractAddress == "" {
		return errors.New("erc20DeployContractAddress is required")
	}

	//使用发币合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(erc20DeployContractAddress), client)
	if err != nil {
		log.Fatalf("conn contract: %v \n", err)
	} else {
		log.Println("NewERC20Token succeed")
	}

	// privateKey, err := crypto.HexToECDSA("fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// auth := bind.NewKeyedTransactor(privateKey)

	var accountBalance = big.NewInt(0)
	if accountBalance, err = contract.BalanceOf(nil, common.HexToAddress(sender)); err != nil {
		log.Fatalf("get Balances err: %v \n", err)
	}
	fmt.Println("发送方LNMC余额: ", accountBalance)

	// target是接收地址
	if accountBalance, err := contract.BalanceOf(nil, common.HexToAddress(receiver)); err != nil {
		log.Fatalf("get Balances err: %v \n", err)
	} else {

		fmt.Println("接收方LNMC余额: ", accountBalance)
	}
	return nil
}

//ERC20代币余额查询， 传参: 账号地址
func queryLNMCBalance(addressHex string) error {

	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	redisConn, err := redis.Dial("tcp", RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()
	//
	erc20DeployContractAddress, _ := redis.String(redisConn.Do("GET", "ERC20DeployContractAddress"))
	if err != nil {
		log.Fatal(err)
		return err
	}
	if erc20DeployContractAddress == "" {
		return errors.New("erc20DeployContractAddress is required")
	}

	//使用发币合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(erc20DeployContractAddress), client)
	if err != nil {
		log.Fatalf("conn contract: %v \n", err)
	} else {
		log.Println("NewERC20Token succeed")
	}

	if accountBalance, err := contract.BalanceOf(nil, common.HexToAddress(addressHex)); err != nil {
		log.Fatalf("get Balances err: %v \n", err)
	} else {

		fmt.Printf("账号[%s]的LNMC余额: %s LNMC\n", addressHex, accountBalance.String())
	}
	return nil

}

//LNMC代币转账, 从sourcePrivateKey对应的地址转到目标账号, amount 以LNMC为单位，每一枚=1分钱
func transferLNMC(sourcePrivateKey, target string, amount int64) error {
	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatal(err)
	}

	redisConn, err := redis.Dial("tcp", RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()
	//
	erc20DeployContractAddress, _ := redis.String(redisConn.Do("GET", "ERC20DeployContractAddress"))
	if err != nil {
		log.Fatal(err)
		return err
	}
	if erc20DeployContractAddress == "" {
		return errors.New("erc20DeployContractAddress is required")
	}

	//使用发币合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(erc20DeployContractAddress), client)
	if err != nil {
		log.Fatalf("conn contract: %v \n", err)
	}

	privateKey, err := crypto.HexToECDSA(sourcePrivateKey) //源钱包地址私钥
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)       // in wei
	auth.GasLimit = uint64(GASLIMIT) // in units
	auth.GasPrice = gasPrice

	//调用合约里的转账函数
	signedTx, err := contract.Transfer(&bind.TransactOpts{
		From:   auth.From,
		Signer: auth.Signer,
		Value:  nil,
	}, common.HexToAddress(target), big.NewInt(amount))
	if err != nil {
		log.Fatalf("TransferFrom err: %v \n", err)
	}

	log.Printf("tx sent: %s\n", signedTx.Hash().Hex())

	done := WaitForBlockCompletation(client, signedTx.Hash().Hex())
	if done == 1 {
		tx2, isPending, err := client.TransactionByHash(context.Background(), signedTx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("代币从5s转到目标账号成功, 交易哈希: %s\n", fromAddress, tx2.Hash().Hex()) //
		log.Println("isPending: ", isPending)                                  // false

	} else {
		log.Println("代币转到目标账号失败")
	}

	return nil
}

/*
多签合约, 从合约账号转到目标账号
传参：
  1. multiSigContractAddress -  第一步部署的多签智能合约， A+B => C
  2. privateKeySource - A或B的私钥，用来签名, 在本系统里，需要派生一个子私钥来作为B(证明人)
  3. target -  目标接收者的地址, C

*/
func transferTokenFromABToC(multiSigContractAddress, privateKeySource, target string, amount int64) {

	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	cAddr := common.HexToAddress(multiSigContractAddress)
	fmt.Println(cAddr.String())

	//调用多签智能合约地址
	contract, err := MultiSig.NewMultiSig(cAddr, client)
	if err != nil {
		log.Fatalf("conn contract: %v \n", err)
	}
	fmt.Println(contract.Name(&bind.CallOpts{Pending: true}))

	privateKey, err := crypto.HexToECDSA(privateKeySource) //A或B私钥
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)       // in wei
	auth.GasLimit = uint64(GASLIMIT) // in units
	auth.GasPrice = gasPrice

	//调用合约里的转账函数, 返回已经签名的tx交易哈希
	transferMultiSigTx, err := contract.Transfer(&bind.TransactOpts{
		From:   auth.From,
		Signer: auth.Signer,
		Value:  nil,
	}, common.HexToAddress(target), big.NewInt(amount)) //LNMC
	if err != nil {
		log.Fatalf("Transfer err: %v \n", err)
	}
	fmt.Printf("tx of multisig contract sent: %s \n", transferMultiSigTx.Hash().Hex())
	fmt.Printf("tx of multisig contract Data bytes: %s \n", transferMultiSigTx.Data())
	fmt.Printf("tx of multisig contract Data hex: %s \n", hex.EncodeToString(transferMultiSigTx.Data()))

	//等待打包完成的回调
	done := WaitForBlockCompletation(client, transferMultiSigTx.Hash().Hex())
	if done == 1 {
		//获取交易哈希里的打包状态，如果打包完成，isPending = false
		tx2, isPending, err := client.TransactionByHash(context.Background(), transferMultiSigTx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		log.Println("multisig contract 打包成功, 交易哈希: ", tx2.Hash().Hex()) //
		log.Println("isPending: ", isPending)                           // false

	} else {
		log.Println("multisig contract 打包失败")
	}

}

//Eth转账, 从sourcePrivateKey转到目标账号target, amount  单位是 wei
func transferEth(sourcePrivateKey, target string, amount string) {
	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := crypto.HexToECDSA(sourcePrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := new(big.Int)
	value.SetString(amount, 10) // in wei sets the value to eth

	//big.NewInt(amount)

	gasLimit := uint64(GASLIMIT) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress(target)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("tx sent: %s\n", signedTx.Hash().Hex())
	//等待打包完成的回调
	done := WaitForBlockCompletation(client, signedTx.Hash().Hex())
	if done == 1 {
		//获取交易哈希里的打包状态，如果打包完成，isPending = false
		tx2, isPending, err := client.TransactionByHash(context.Background(), signedTx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		log.Println("打包成功, 交易哈希: ", tx2.Hash().Hex()) //
		log.Println("isPending: ", isPending)         // false

	} else {
		log.Println(" 打包失败")
	}

}

/*
通过构造裸交易数据进行代币的转账
*/
func transferToken() error {
	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatal(err)
	}
	redisConn, err := redis.Dial("tcp", RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()
	//
	erc20DeployContractAddress, _ := redis.String(redisConn.Do("GET", "ERC20DeployContractAddress"))
	if err != nil {
		log.Fatal(err)
		return err
	}
	if erc20DeployContractAddress == "" {
		return errors.New("erc20DeployContractAddress is required")
	}

	//约定，使用第1号叶子发币
	privateKey, err := crypto.HexToECDSA("fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("nonce:", int64(nonce))

	value := big.NewInt(0) // in wei (0 eth) 由于进行的是代币转账，不设计以太币转账，因此这里填0
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("gasPrice", gasPrice)

	//接收者地址：用户D
	toAddress := common.HexToAddress("0x59aC768b416C035c8DB50B4F54faaa1E423c070D")

	//注意，这里需要填写发币的智能合约地址
	tokenAddress := common.HexToAddress(erc20DeployContractAddress)

	// internal/pkg/blockchain/lnmc/contracts/ERC20/ERC20Token.sol 第54行
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress)) // 0x00000000000000000000000059ac768b416c035c8db50b4f54faaa1e423c070d

	amount := new(big.Int)
	amount.SetString("100", 10) // 100 tokens

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x0000000000000000000000000000000000000000000000000000000000000064

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	fmt.Println("data:", data)
	fmt.Println("data hex:", hex.EncodeToString(data))
	//a9059cbb00000000000000000000000059ac768b416c035c8db50b4f54faaa1e423c070d0000000000000000000000000000000000000000000000000000000000000064

	// gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
	// 	To:   &toAddress,
	// 	Data: data,
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("gasLimit:", gasLimit) // 21572

	//注意！！！不能用上面的，否则无法打包
	gasLimit := uint64(GASLIMIT) // in units

	//构造代币转账的交易裸数据
	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	//对裸交易数据签名
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("transferToken, tx sent: %s\n", signedTx.Hash().Hex())

	//等待打包完成的回调
	done := WaitForBlockCompletation(client, signedTx.Hash().Hex())
	if done == 1 {
		//获取交易哈希里的打包状态，如果打包完成，isPending = false
		tx2, isPending, err := client.TransactionByHash(context.Background(), signedTx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		log.Println("打包成功, 交易哈希: ", tx2.Hash().Hex()) //
		log.Println("isPending: ", isPending)         // false

	} else {
		log.Println(" 打包失败")
	}
	return nil
}

/*
通过构造MultiSig  合约  裸交易数据进行代币的转账
contractAddress - 多签合约地址
privKey - A 或 B 私钥
target - 接收者账号
*/
func transferMultiSigToken(contractAddress, privKey, target string, tokens string) {
	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatal(err)
	}

	//A 或 B 私钥
	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("nonce:", int64(nonce))

	value := big.NewInt(0) // in wei (0 eth) 由于进行的是代币转账，不设计以太币转账，因此这里填0
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("gasPrice", gasPrice)

	//接收者地址：用户D
	toAddress := common.HexToAddress(target)

	//注意，这里需要填写多签合约地址
	tokenAddress := common.HexToAddress(contractAddress)

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress)) // 0x00000000000000000000000059ac768b416c035c8db50b4f54faaa1e423c070d

	amount := new(big.Int)
	amount.SetString(tokens, 10) //代币数量

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	fmt.Println("data:", data)
	fmt.Println("data hex:", hex.EncodeToString(data))

	gasLimit := uint64(GASLIMIT) //必须强行指定，否则无法打包
	fmt.Println("gasLimit:", gasLimit)

	//构造代币转账的交易裸数据
	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("chainID:", chainID.String())

	//对裸交易数据签名
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("transferToken, tx sent: %s\n", signedTx.Hash().Hex())

	//等待打包完成的回调
	done := WaitForBlockCompletation(client, signedTx.Hash().Hex())
	if done == 1 {
		//获取交易哈希里的打包状态，如果打包完成，isPending = false
		tx2, isPending, err := client.TransactionByHash(context.Background(), signedTx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		log.Println("打包成功, 交易哈希: ", tx2.Hash().Hex()) //
		log.Println("isPending: ", isPending)         // false

	} else {
		log.Println(" 打包失败")
	}

}

//

func GenerateRawDesc(contractAddress, fromAddressHex, target, tokens string) ([]byte, error) {
	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatal(err)
	}

	fromAddress := common.HexToAddress(fromAddressHex)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("nonce:", int64(nonce))

	value := big.NewInt(0) // in wei (0 eth) 由于进行的是代币转账，不设计以太币转账，因此这里填0
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("gasPrice", gasPrice)

	//接收者地址：用户D
	toAddress := common.HexToAddress(target)

	//注意，这里需要填写多签合约地址
	tokenAddress := common.HexToAddress(contractAddress)

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress)) // 0x00000000000000000000000059ac768b416c035c8db50b4f54faaa1e423c070d

	amount := new(big.Int)
	amount.SetString(tokens, 10) //代币数量

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	fmt.Println("data:", data)
	fmt.Println("data hex:", hex.EncodeToString(data))

	gasLimit := uint64(GASLIMIT) //必须强行指定，否则无法打包
	fmt.Println("gasLimit:", gasLimit)

	//构造代币转账的交易裸数据
	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("chainID:", chainID.String())
	_ = chainID

	return tx.MarshalJSON()
}

/*
一个普通用户账号转账，目标是多签合约地址
*/
func transferLNMCTokenToContractAddress(privKeyHex, target, tokens string) ([]byte, error) {
	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatal(err)
	}

	redisConn, err := redis.Dial("tcp", RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	defer redisConn.Close()
	//
	erc20DeployContractAddress, _ := redis.String(redisConn.Do("GET", "ERC20DeployContractAddress"))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	if erc20DeployContractAddress == "" {
		return nil, errors.New("erc20DeployContractAddress is required")
	}

	//A私钥
	privateKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("nonce:", int64(nonce))

	value := big.NewInt(0) // in wei (0 eth) 由于进行的是代币转账，不设计以太币转账，因此这里填0
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("gasPrice", gasPrice)

	//接收者地址：多签合约地址
	toAddress := common.HexToAddress(target)

	//注意，这里需要填写发币合约地址
	tokenAddress := common.HexToAddress(erc20DeployContractAddress)

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress))

	amount := new(big.Int)
	amount.SetString(tokens, 10) //代币数量

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount))

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	fmt.Println("data:", data)
	fmt.Println("data hex:", hex.EncodeToString(data))

	gasLimit := uint64(GASLIMIT) //必须强行指定，否则无法打包
	fmt.Println("gasLimit:", gasLimit)

	//构造代币转账的交易裸数据
	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("chainID:", chainID.String())

	//对裸交易数据签名
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("transferToken, tx sent: %s\n", signedTx.Hash().Hex())

	//等待打包完成的回调
	done := WaitForBlockCompletation(client, signedTx.Hash().Hex())
	if done == 1 {
		//获取交易哈希里的打包状态，如果打包完成，isPending = false
		tx2, isPending, err := client.TransactionByHash(context.Background(), signedTx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		log.Println("打包成功, 交易哈希: ", tx2.Hash().Hex()) //
		log.Println("isPending: ", isPending)         // false

	} else {
		log.Println(" 打包失败")
	}

	return nil, nil
}

/*
构造一个普通用户账号转账的裸交易数据，目标是多签合约地址
source - 发起方账号
target - 多签合约地址
tokens - 代币数量，字符串格式
*/
func generateTransferLNMCTokenTx(source, target, tokens string) (*RawDesc, error) {
	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatal(err)
	}

	redisConn, err := redis.Dial("tcp", RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	defer redisConn.Close()
	//
	erc20DeployContractAddress, _ := redis.String(redisConn.Do("GET", "ERC20DeployContractAddress"))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	if erc20DeployContractAddress == "" {
		return nil, errors.New("erc20DeployContractAddress is required")
	}

	fromAddress := common.HexToAddress(source)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("nonce:", int64(nonce))

	value := big.NewInt(0) // in wei (0 eth) 由于进行的是代币转账，不设计以太币转账，因此这里填0
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("gasPrice", gasPrice)

	//接收者地址：多签合约地址
	toAddress := common.HexToAddress(target)

	//注意，这里需要填写发币合约地址
	tokenAddress := common.HexToAddress(erc20DeployContractAddress)

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress))

	amount := new(big.Int)
	amount.SetString(tokens, 10) //代币数量

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount))

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	fmt.Println("data:", data)
	fmt.Println("data hex:", hex.EncodeToString(data))

	gasLimit := uint64(GASLIMIT) //必须强行指定，否则无法打包
	fmt.Println("gasLimit:", gasLimit)

	//构造代币转账的交易裸数据
	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("chainID:", chainID.String())

	_ = tx

	return &RawDesc{
		Nonce:           nonce,
		GasPrice:        gasPrice.Uint64(),
		GasLimit:        gasLimit,
		ChainID:         chainID.Uint64(),
		Txdata:          data,
		ContractAddress: target, //合约地址
		Value:           0,
	}, nil
}

//传入Rawtx， 进行签名, 构造一个已经签名的hex裸交易
func buildTx(rawDesc *RawDesc, privKeyHex, contractAddress string) (string, error) {

	//A私钥
	privateKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		log.Fatal(err)
	}

	//注意，这里需要填写发币合约地址，不能设为刚刚创建的合约地址
	// tokenAddress := common.HexToAddress(ERC20DeployContractAddress)
	tokenAddress := common.HexToAddress(contractAddress)

	//构造代币转账的交易裸数据
	tx := types.NewTransaction(
		rawDesc.Nonce,
		tokenAddress, //to
		big.NewInt(int64(rawDesc.Value)),
		rawDesc.GasLimit,
		big.NewInt(int64(rawDesc.GasPrice)),
		rawDesc.Txdata,
	)

	//对裸交易数据签名
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(int64(rawDesc.ChainID))), privateKey)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	rawTxHex := hex.EncodeToString(rawTxBytes)

	log.Println("rawTxHex:", rawTxHex)
	return rawTxHex, nil
}

//根据客户端SDK签名后的裸交易数据，广播到链上
// rawTxHex := "f86d8202b28477359400825208944592d8f8d7b001e72cb26a73e4fa1806a51ac79d880de0b6b3a7640000802ca05924bde7ef10aa88db9c66dd4f5fb16b46dff2319b9968be983118b57bb50562a001b24b31010004f13d9a26b320845257a6cfc2bf819a3d55e3fc86263c5f0772"
func sendSignedTxToGeth(rawTxHex string) error {
	client, err := ethclient.Dial(WSURIIPC)
	if err != nil {
		log.Fatal(err)
	}

	rawTxBytes, err := hex.DecodeString(rawTxHex)

	signedTx := new(types.Transaction)
	rlp.DecodeBytes(rawTxBytes, &signedTx)

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
		return err
	}

	fmt.Printf("signedTx sent: %s", signedTx.Hash().Hex())

	//等待打包完成的回调
	done := WaitForBlockCompletation(client, signedTx.Hash().Hex())
	if done == 1 {
		//获取交易哈希里的打包状态，如果打包完成，isPending = false
		tx2, isPending, err := client.TransactionByHash(context.Background(), signedTx.Hash())
		if err != nil {
			log.Fatal(err)
			return err
		}

		log.Println("打包成功, 交易哈希: ", tx2.Hash().Hex()) //
		log.Println("isPending: ", isPending)         // false

	} else {
		log.Println(" 打包失败")
	}
	return nil
}
func mulsigDeployMain() {

	//第一步: 部署 多签合约 使用第1片叶子部署合约
	// deployMultiSig("fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915")
	/*
		Contract pending deploy: 0x3Eb7A38688e6805DA14c02F1aE925a85562367C7, SigTx Hash: 0x8baa8843fafd1d18ce3ab66b15a41bb84bac2ed3e40dfa66f0f7ecc590b5ae52

	*/

	// 第二步:  从A账户将若干代币转账到刚刚部署的多签合约
	// sendTokenToMultisigContractAddress(PrivateKeyAHEX, "0x3Eb7A38688e6805DA14c02F1aE925a85562367C7", "50")
	// transferLNMCTokenToContractAddress(PrivateKeyAHEX, "0x3Eb7A38688e6805DA14c02F1aE925a85562367C7", "50")

	// 第三步: A调用智能合约进行转账
	// transferTokenFromABToC(
	// 	"0x3Eb7A38688e6805DA14c02F1aE925a85562367C7",
	// 	PrivateKeyAHEX, //A私钥
	// 	AddressDHEX,    //D账号地址
	// 	50)
	//交易哈希 tx: 0x4ccbb25233451a5d4b47dca0f9517a38a8b25b264b9bf5de853d777db771f527

	// 第四步:  B审核
	// transferTokenFromABToC(
	// 	"0x3Eb7A38688e6805DA14c02F1aE925a85562367C7",
	// 	PrivateKeyBHEX, //B私钥
	// 	AddressDHEX,    //D账号地址
	// 	50)

	//查询用户C的余额
	// queryLNMCBalance(AddressCHEX)

}

func main() {
	//部署合约于第1号叶子 ,  获取发币合约地址
	if err := deploy("fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915"); err != nil {
		log.Println(err.Error())
	}
	//输出: Contract pending deploy:0x1d2bdda8954b401feb52008c63878e698b6b8444

	//查询第1号叶子的LNMC余额
	getTokenBalance("0x4acea697f366C47757df8470e610a2d9B559DbBE")
	//输出: Token of LNMC: 10000000000

	//从第1号叶子转账 1000000000000000000 wei到id2
	transferEth("fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915", "0x9d8D057020C6d5e2994520a74298ACB80aAdDB55", "1000000000000000000")

	//从第1号叶子转账500代币给A
	// transferLNMC("fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915", AddressAHEX, 200)

	//从第1号叶子转账5200代币给id4
	// transferLNMC("fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915", "0x9858effd232b4033e47d90003d41ec34ecaeda94", 300)

	//查询第1号叶子余额，约定为第1号叶子的地址 用于验证
	queryLNMCBalance("0x4acea697f366C47757df8470e610a2d9B559DbBE")
	//查询id4代币余额
	// queryLNMCBalance("0x9858effd232b4033e47d90003d41ec34ecaeda94")
	//查询用户A的余额
	// queryLNMCBalance(AddressAHEX)
	//查询用户B的余额
	// queryLNMCBalance(AddressBHEX)

	fmt.Println("========")

	// re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	// addressHex := "0x9d8D057020C6d5e2994520a74298ACB80aAdDB55"
	// if re.MatchString(addressHex) == true {
	// 	fmt.Println("address is valid")
	// } else {
	// 	fmt.Println("address is not valid")
	// }

	// //查询新注册用户的ETH余额
	// queryLNMCBalance(AddressCHEX)

	//   代币转账
	// transferToken()

	// 模拟客户端A签
	// transferMultiSigToken(
	// 	"0x3Eb7A38688e6805DA14c02F1aE925a85562367C7",
	// 	PrivateKeyAHEX, //A私钥
	// 	AddressDHEX,    //D账号地址
	// 	"50",
	// )

	// 模拟 服务端B签 审核
	// transferMultiSigToken(
	// 	"0x3Eb7A38688e6805DA14c02F1aE925a85562367C7",
	// 	PrivateKeyBHEX, //B私钥
	// 	AddressDHEX,    //D账号地址
	// 	"50",
	// )

	// 从第1号叶子转账10代币给D
	// transferLNMC("fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915", "0x59aC768b416C035c8DB50B4F54faaa1E423c070D", 10)

	/*
		rawTx, err := GenerateRawTx(
			"0x3Eb7A38688e6805DA14c02F1aE925a85562367C7",
			AddressAHEX, //from
			AddressDHEX, //D账号地址
			"50",
		)
		if err != nil {
			fmt.Println("GenerateRawTx error :", err)

			return
		}
		log.Println("rawTx:", rawTx)
	*/

	//查询用户D的余额
	// queryLNMCBalance("0x59aC768b416C035c8DB50B4F54faaa1E423c070D")

	/*
		//构造普通用户转账到多签合约的裸交易数据
		rawDesc, err := generateTransferLNMCTokenTx(AddressAHEX, "0x3Eb7A38688e6805DA14c02F1aE925a85562367C7", "50")
		if err != nil {
			log.Fatalln(err)

		}

		//模仿SDK，进行签名，注意：第三个参数必须是erc20发币合约地址
		rawTxHex, err := buildTx(rawDesc, PrivateKeyAHEX, ERC20DeployContractAddress)
		if err != nil {
			log.Fatalln(err)

		}
		err = sendSignedTxToGeth(rawTxHex)
		if err != nil {
			log.Fatalln(err)

		}
		//查询刚刚部署多签合约的余额， 应该是150
		queryLNMCBalance("0x3Eb7A38688e6805DA14c02F1aE925a85562367C7")
	*/

	//查询id2的多签合约的余额
	// queryLNMCBalance("0x2a2D3d88f9F385559CC7C3283Be95E3CcaE6029E")

}

/*
eth.sendTransaction({from:account1,to:account3,value:web3.toWei(1,"ether")})
*/
