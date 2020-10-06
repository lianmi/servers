package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	// "math"
	// "os"
	// "strings"
	"crypto/ecdsa"
	"github.com/pkg/errors"
	// "github.com/ethereum/go-ethereum/rpc"
	"math/big"
	// "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/lianmi/servers/internal/pkg/blockchain/util"
	"github.com/miguelmota/go-ethereum-hdwallet"
	"io/ioutil"
)

const (
	KEY             = "UTC--2020-10-06T16-30-19.524731110Z--7562b4d3b08b2373e68d4e89f69f6fb731b308e1"
	COINBASEACCOUNT = "0x7562b4d3b08b2373e68d4e89f69f6fb731b308e1"
	PASSWORD        = "LianmiSky8900388"
	GASLIMIT        = 6000000
)

func createHDWallet() {
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

// GetLatestBlockNumber get the latest block number
func getLatestBlockNumber() (*big.Int, error) {
	client, err := ethclient.Dial("ws://127.0.0.1:8546")
	if err != nil {
		log.Fatal(err)
	}
	block, err := client.HeaderByNumber(context.Background(), nil)
	return block.Number, err
}

// GetPublicAddressFromPrivateKey returns public address from private key
func getPublicAddressFromPrivateKey(priv *ecdsa.PrivateKey) (common.Address, error) {
	var address common.Address
	pub := priv.Public()
	pubECDSA, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return address, errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	address = crypto.PubkeyToAddress(*pubECDSA)
	return address, nil
}

// GetGasPrice gets clamped gas price
func getGasPrice() *big.Int {
	client, err := ethclient.Dial("ws://127.0.0.1:8546")
	if err != nil {
		log.Fatal(err)
	}
	maxGasPrice := big.NewInt(9000000000)     // 9 gwei
	defaultGasPrice := big.NewInt(1000000000) // 1 gwei
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return defaultGasPrice
	}

	// cap gas price in case SuggestGasPrice goes off the rails
	if gasPrice.Cmp(maxGasPrice) == 1 {
		return maxGasPrice
	}

	return gasPrice
}

// 输出wei为单位的账户余额
func getWeiBalance(address string) *big.Int {
	client, err := ethclient.Dial("ws://127.0.0.1:8546")
	// client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatal(err)
	}

	account := common.HexToAddress(address)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("balance: ", balance)
	return balance

}

// 输出Eth为单位的账户余额
func getEthBalance(address string) float64 {
	// client, err := ethclient.Dial("http://127.0.0.1:8545")
	client, err := ethclient.Dial("ws://127.0.0.1:8546")
	if err != nil {
		log.Fatal(err)
	}

	account := common.HexToAddress(address)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Fatal(err)
	}
	ethAmount := util.ToDecimal(balance, 18)
	f64, _ := ethAmount.Float64()
	fmt.Println("balance(Eth): ", f64)
	return f64
}

func checkAccountIsvalid() {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")

	fmt.Printf("is valid: %v\n", re.MatchString("0x323b5d4c32345ced77393b3530b1eed0f346429d")) // is valid: true
	fmt.Printf("is valid: %v\n", re.MatchString("0xZYXb5d4c32345ced77393b3530b1eed0f346429d")) // is valid: false

	// client, err := ethclient.Dial("http://127.0.0.1:8545")
	client, err := ethclient.Dial("ws://127.0.0.1:8546")
	if err != nil {
		log.Fatal(err)
	}

	// 0x Protocol Token (ZRX) smart contract address
	address := common.HexToAddress("0xe41d2489571d322189246dafa5ebde1f4699f498")
	bytecode, err := client.CodeAt(context.Background(), address, nil) // nil is latest block
	if err != nil {
		log.Fatal(err)
	}

	isContract := len(bytecode) > 0

	fmt.Printf("is contract: %v\n", isContract) // is contract: true

	// a random user account address
	address = common.HexToAddress("0x8e215d06ea7ec1fdb4fc5fd21768f4b34ee92ef4")
	bytecode, err = client.CodeAt(context.Background(), address, nil) // nil is latest block
	if err != nil {
		log.Fatal(err)
	}

	isContract = len(bytecode) > 0

	fmt.Printf("is contract: %v\n", isContract) // is contract: false
}

// 根据keystore文件与密码生成私钥
func KeystoreToPrivateKey(privateKeyFile, password string) (string, string, error) {
	keyjson, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		fmt.Println("read keyjson file failed：", err)
	}
	unlockedKey, err := keystore.DecryptKey(keyjson, password)
	if err != nil {

		return "", "", err

	}
	privKey := hex.EncodeToString(unlockedKey.PrivateKey.D.Bytes())
	addr := crypto.PubkeyToAddress(unlockedKey.PrivateKey.PublicKey)
	return privKey, addr.String(), nil
}

//检查交易打包状态
func checkTransactionReceipt(_txHash string) int {
	// client, err := ethclient.Dial("http://127.0.0.1:8545")
	client, err := ethclient.Dial("ws://127.0.0.1:8546")
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
			_ = err
			return -1
		case header := <-headers:
			log.Println(header.TxHash.Hex())
			transactionStatus := checkTransactionReceipt(hashToRead)
			if transactionStatus == 0 {
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
	client, err := ethclient.Dial("ws://127.0.0.1:8546")
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
		log.Println("tx.Hash: ", tx.Hash().Hex())      //
		log.Println("tx.Value: ", tx.Value().String()) // 10000000000000000
		log.Println("tx.Gas: ", tx.Gas())
		log.Println("tx.GasPrice: ", tx.GasPrice().Uint64()) // 1000000000

		// cost := tx.Gas() * tx.GasPrice().Uint64() //计算交易所需要支付的总费用
		gasCost := util.CalcGasCost(tx.Gas(), tx.GasPrice()) //计算交易所需要支付的总费用
		log.Println("交易总费用(Wei): ", gasCost)                 //6000000000000000Wei
		ethAmount := util.ToDecimal(gasCost, 18)
		log.Println("交易总费用(eth): ", ethAmount) //0.003Eth
		log.Println("tx.Nonce: ", tx.Nonce())
		log.Println("tx.Data: ", tx.Data())
		log.Println("tx.To: ", tx.To().Hex()) //目标地址

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

//从总账号地址转账Eth到其它普通账号地址, 以wei为单位, 1 eth = 1x18次方wei amount 是字符型
func transferEthFromCoinbaseToOtherAccount(targetAccount string, amount string) {
	// client, err := ethclient.Dial("http://127.0.0.1:8545")
	client, err := ethclient.Dial("ws://127.0.0.1:8546")
	if err != nil {
		log.Fatal(err)
	}

	privKeyHex, addressHex, err := KeystoreToPrivateKey(KEY, PASSWORD)
	if err != nil {
		log.Fatal(err)

	}
	fmt.Printf("privKeyHex: %s\n address: %s\n", privKeyHex, addressHex)
	//privKeyHex: bc2f812f1f534c9e8a3b3cfb628b0ea5d41967d4f18391c6489737d743b1ee7a
	//address: 0xB18Db89641D2ec807104258e2205e6AC6264BF25

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

	value := new(big.Int)
	value.SetString(amount, 10) // in wei sets the value to eth

	gasLimit := uint64(GASLIMIT) // in units

	gasPrice := getGasPrice()

	//接收账号
	toAddress := common.HexToAddress(targetAccount)

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

	fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex())

	/*
		等待检测交易是否完成，挖矿工需要工作才能出块
		> miner.start()
		> var account2="0x4acea697f366C47757df8470e610a2d9B559DbBE"
		> web3.fromWei(web3.eth.getBalance(account2), 'ether')
		输出： 1
	*/

	done := WaitForBlockCompletation(client, signedTx.Hash().Hex())
	if done == 1 {
		log.Println("交易完成")
		tx, isPending, err := client.TransactionByHash(context.Background(), signedTx.Hash())
		if err != nil {
			log.Fatal(err)
		}
		// txHash := common.HexToHash(tx.Hash().Hex())
		log.Println("交易哈希: ", tx.Hash().Hex()) //
		log.Println("isPending: ", isPending)  // false
	} else {
		log.Println("交易失败")
	}

}

//从第0号叶子地址转账Eth到其它普通账号地址, 以wei为单位, 1 eth = 1x18次方wei amount 是字符型
func transferEthFromLeaf0ToOtherAccount(targetAccount string, amount string) {
	client, err := ethclient.Dial("ws://127.0.0.1:8546")
	if err != nil {
		log.Fatal(err)
	}

	//第0号叶子私钥
	privKeyHex := "4c88e6ccffec59b6c3df5ab51a4e6c42c421f58274d653d716aafd4aff376f5b"

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

	value := new(big.Int)
	value.SetString(amount, 10) // sets the value to eth

	gasLimit := uint64(GASLIMIT) // in units

	gasPrice := getGasPrice()

	//接收账号
	toAddress := common.HexToAddress(targetAccount)

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

	fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex())

	/*
		等待检测交易是否完成，挖矿工需要工作才能出块
		> miner.start()
		> var account2="0x4acea697f366C47757df8470e610a2d9B559DbBE"
		> web3.fromWei(web3.eth.getBalance(account2), 'ether')
		输出： 1
	*/

	done := WaitForBlockCompletation(client, signedTx.Hash().Hex())
	if done == 1 {
		log.Println("交易完成")
		tx, isPending, err := client.TransactionByHash(context.Background(), signedTx.Hash())
		if err != nil {
			log.Fatal(err)
		}
		// txHash := common.HexToHash(tx.Hash().Hex())
		log.Println("交易哈希: ", tx.Hash().Hex()) //
		log.Println("isPending: ", isPending)  // false
	} else {
		log.Println("交易失败")
	}

}

func main() {
	// 以wei为单位输出某个地址的eth
	getWeiBalance(COINBASEACCOUNT)

	// getEthBalance("0xb18db89641d2ec807104258e2205e6ac6264bf25")

	//从挖矿账号转账到第0号叶子
	transferEthFromCoinbaseToOtherAccount("0xe14D151e0511b61357DDe1B35a74E9c043c34C47", "994000000000000000000")

	//从第0号叶子向普通用户账号A转eth 1eth
	// transferEthFromLeaf0ToOtherAccount("0x6d9CFbC20E1b210d25b84F83Ba546ea4264DA538", "1000000000000000000")

	//从第0号叶子向普通用户账号B转eth 1eth
	// transferEthFromLeaf0ToOtherAccount("0xac243c2FED19d085bF682d0D74e677c1d9911e83", "1000000000000000000")

	number, _ := getLatestBlockNumber()
	log.Println("number", number)
}
