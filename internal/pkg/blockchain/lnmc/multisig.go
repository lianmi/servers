package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/lianmi/servers/internal/pkg/blockchain/util"
	// "io/ioutil"
	// "strings"
	ERC20 "github.com/lianmi/servers/internal/pkg/blockchain/lnmc/contracts/ERC20"
	MultiSig "github.com/lianmi/servers/internal/pkg/blockchain/lnmc/contracts/MultiSig"
	"regexp"
)

const (
	ERC20DeployContractAddress = "0x23a9497bb4ffa4b9d97d3288317c6495ecd3a2ce"                       // ERC20发币地址
	PrivateKeyAHEX             = "91e5f2d81444905af5f94d6b36be36d69363420b9edd59808caec17830d50ff1" //用户A私钥
	AddressAHEX                = "0x6d9CFbC20E1b210d25b84F83Ba546ea4264DA538"                       //用户A地址

	PrivateKeyBHEX = "b65e1f6e3b449c35c18518cfdf8de3c361ccf6f4a51817e0709a917fac688423" //用户B私钥
	AddressBHEX    = "0xac243c2FED19d085bF682d0D74e677c1d9911e83"                       //用户B地址

	AddressCHEX = "0xBa8d69ba4D65802039cfE2ae373072639026D457" //用户C地址
)

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
		log.Println("tx.Hash: ", tx.Hash().Hex())            //
		log.Println("tx.Value: ", tx.Value().String())       // 10000000000000000
		log.Println("tx.Gas: ", tx.Gas())                    // 3000000
		log.Println("tx.GasPrice: ", tx.GasPrice().Uint64()) // 1000000000

		// cost := tx.Gas() * tx.GasPrice().Uint64() //计算交易所需要支付的总费用
		gasCost := util.CalcGasCost(tx.Gas(), tx.GasPrice()) //计算交易所需要支付的总费用
		log.Println("交易总费用(Wei): ", gasCost)                 //3000000000000000Wei
		ethAmount := util.ToDecimal(gasCost, 18)
		log.Println("交易总费用(eth): ", ethAmount) //0.003Eth
		log.Println("tx.Nonce: ", tx.Nonce())
		log.Println("tx.Data: ", tx.Data())
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
func deployMultiSig(privateKeyHex string) {
	client, err := ethclient.Dial("ws://127.0.0.1:8546")
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
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
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(3000000) // in units
	auth.GasPrice = gasPrice

	//用户A的私钥
	privateKeyA, err := crypto.HexToECDSA(PrivateKeyAHEX)
	if err != nil {
		log.Fatal(err)
	}
	publicKeyA := privateKeyA.Public()
	publicKeyECDSAA, ok := publicKeyA.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddressA := crypto.PubkeyToAddress(*publicKeyECDSAA)

	// 商户B的私钥
	privateKeyB, err := crypto.HexToECDSA(PrivateKeyBHEX)
	if err != nil {
		log.Fatal(err)
	}
	publicKeyB := privateKeyB.Public()
	publicKeyECDSAB, ok := publicKeyB.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddressB := crypto.PubkeyToAddress(*publicKeyECDSAB)
	fmt.Println("fromAddressA: ", fromAddressA.String(), "fromAddressB: ", fromAddressB.String())
	//
	address, deployMultiSigTx, _, err := MultiSig.DeployMultiSig(
		auth,
		client,
		fromAddressA, //A 账号地址
		fromAddressB, //B 账号地址
		common.HexToAddress(ERC20DeployContractAddress), //ERC20发币地址
	)
	if err != nil {
		log.Fatalf("deploy %v \n", err)
	}
	fmt.Printf("Contract pending deploy: %s, SigTx Hash: %s\n", address.String(), deployMultiSigTx.Hash().String())

	//TODO 监听，直到合约部署成功,如果失败，则提示

	done := WaitForBlockCompletation(client, deployMultiSigTx.Hash().Hex())
	if done == 1 {
		log.Println("交易完成")
		tx, isPending, err := client.TransactionByHash(context.Background(), deployMultiSigTx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		log.Println("多签合约部署成功, 交易哈希: ", tx.Hash().Hex()) //
		log.Println("isPending: ", isPending)            // false

	} else {
		log.Println("多签合约部署失败")
	}
}

//将一定数量amount的代币由 总发币合约地址 转账到多签合约账户, soure是用户A
func sendTokenToMultisigContractAddress(sourcePrivateKey, target string, amount string) {
	client, err := ethclient.Dial("ws://127.0.0.1:8546")
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	//使用总发币的合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(ERC20DeployContractAddress), client)
	if err != nil {
		log.Fatalf("conn contract: %v \n", err)
	}

	//A的私钥
	privateKey, err := crypto.HexToECDSA(sourcePrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	auth := bind.NewKeyedTransactor(privateKey)

	value := new(big.Int)
	value.SetString(amount, 10) // in wei sets the value to eth

	//调用合约里的转账函数
	transferTx, err := contract.Transfer(&bind.TransactOpts{
		From:   auth.From,
		Signer: auth.Signer,
		Value:  nil,
	}, common.HexToAddress(target), value)
	if err != nil {
		log.Fatalf("TransferFrom err: %v \n", err)
	}
	fmt.Printf("tx sent: %s \n", transferTx.Hash().Hex())

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
}

//ERC20代币余额查询， 传参1是发送者合约地址，传参2是接收者账号地址
func querySendAndReceive2(sender, receiver string) {

	client, err := ethclient.Dial("ws://127.0.0.1:8546")
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	//使用发币合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(ERC20DeployContractAddress), client)
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

}

//ERC20代币余额查询， 传参: 账号地址
func queryLNMCBalance(addressHex string) {

	client, err := ethclient.Dial("ws://127.0.0.1:8546")
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	//使用发币合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(ERC20DeployContractAddress), client)
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

}

//LNMC代币转账, 从第1号叶子转到目标账号, amount 以LNMC为单位，每一枚=1分钱
func transferLNMC(sourcePrivateKey, target string, amount int64) {
	client, err := ethclient.Dial("ws://127.0.0.1:8546")
	if err != nil {
		log.Fatal(err)
	}

	//使用发币合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(ERC20DeployContractAddress), client)
	if err != nil {
		log.Fatalf("conn contract: %v \n", err)
	}

	privateKey, err := crypto.HexToECDSA(sourcePrivateKey) //第1号叶子私钥
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
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(3000000) // in units
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

	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())

	done := WaitForBlockCompletation(client, signedTx.Hash().Hex())
	if done == 1 {
		tx2, isPending, err := client.TransactionByHash(context.Background(), signedTx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		log.Println("代币从第1号叶子转到目标账号成功, 交易哈希: ", tx2.Hash().Hex()) //
		log.Println("isPending: ", isPending)                     // false

	} else {
		log.Println("代币从第1号叶子转到目标账号失败")
	}
}

/*
多签合约, 从合约账号转到目标账号
传参：
  1. multiSigContractAddress 第一步部署的多签智能合约， A+B => C
  2. privateKeySource A或B的私钥，用来签名
  3. target 目标地址C, 在本系统里，需要派生一个子地址来接收

*/
func transferTokenFromABToC(multiSigContractAddress, privateKeySource, target string, amount int64) {

	client, err := ethclient.Dial("ws://127.0.0.1:8546")
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
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(3000000) // in units
	auth.GasPrice = gasPrice

	//调用合约里的转账函数
	transferMultiSigTx, err := contract.Transfer(&bind.TransactOpts{
		From:   auth.From,
		Signer: auth.Signer,
		Value:  nil,
	}, common.HexToAddress(target), big.NewInt(amount)) //LNMC
	if err != nil {
		log.Fatalf("Transfer err: %v \n", err)
	}
	fmt.Printf("tx of multisig contract sent: %s \n", transferMultiSigTx.Hash().Hex())

	done := WaitForBlockCompletation(client, transferMultiSigTx.Hash().Hex())
	if done == 1 {
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

func main() {
	//第一步: 部署 多签合约 使用第1片叶子部署合约
	// deployMultiSig("fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915")
	/*
		Contract pending deploy: 0x9696964DC575Eb1Ae9137e6DD9D068307BA569F1, SigTx Hash: 0x8baa8843fafd1d18ce3ab66b15a41bb84bac2ed3e40dfa66f0f7ecc590b5ae52

	*/

	//从第1号叶子转账500代币给A
	// transferLNMC("fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915", AddressAHEX, 500)

	// 第二步:  从A账户将若干代币转账到刚刚部署的多签合约
	// sendTokenToMultisigContractAddress(PrivateKeyAHEX, "0x9696964DC575Eb1Ae9137e6DD9D068307BA569F1", "50")

	//查询第1号叶子余额，约定为第1号叶子的地址 用于验证
	queryLNMCBalance("0x4acea697f366C47757df8470e610a2d9B559DbBE")
	//查询刚刚部署多签合约的余额， 应该是50
	queryLNMCBalance("0x9696964DC575Eb1Ae9137e6DD9D068307BA569F1")
	//查询用户A的余额
	queryLNMCBalance(AddressAHEX)
	//查询用户B的余额
	queryLNMCBalance(AddressBHEX)

	fmt.Println("========")

	// 第三步: A调用智能合约进行转账
	// transferTokenFromABToC(
	// 	"0x9696964DC575Eb1Ae9137e6DD9D068307BA569F1",
	// 	PrivateKeyAHEX, //A私钥
	// 	AddressCHEX,    //C账号地址
	// 	50)

	// 第四步:  B审核
	// transferTokenFromABToC(
	// 	"0x9696964DC575Eb1Ae9137e6DD9D068307BA569F1",
	// 	PrivateKeyBHEX, //B私钥
	// 	AddressCHEX,    //C账号地址
	// 	50)

	//查询用户C的余额
	queryLNMCBalance(AddressCHEX)

	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	addressHex := "0x9d8D057020C6d5e2994520a74298ACB80aAdDB55"
	if re.MatchString(addressHex) == true {
		fmt.Println("address is valid")
	} else {
		fmt.Println("address is not valid")
	}

	//查询新注册用户的ETH余额
	queryLNMCBalance(AddressCHEX)

}
