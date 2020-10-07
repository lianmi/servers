package main

import (
	"context"
	"crypto/ecdsa"

	"fmt"

	// "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	ERC20 "github.com/lianmi/servers/internal/pkg/blockchain/lnmc/contracts/ERC20"
	// "io/ioutil"
	"log"
	"math/big"
	// "strings"
)

const (
	WSURI           = "ws://172.17.0.1:8546" //"ws://127.0.0.1:8546"
	KEY             = "UTC--2020-10-06T16-30-19.524731110Z--7562b4d3b08b2373e68d4e89f69f6fb731b308e1"
	COINBASEACCOUNT = "0x7562b4d3b08b2373e68d4e89f69f6fb731b308e1"
	PASSWORD        = "LianmiSky8900388"
	GASLIMIT        = 6000000
)

/*
发币合约部署
第1号叶子:
privateKey: fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915
address: 0x4acea697f366C47757df8470e610a2d9B559DbBE
*/

func deploy(privateKeyHex string) {
	client, err := ethclient.Dial(WSURI)
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
	auth.Value = big.NewInt(0)       // in wei
	auth.GasLimit = uint64(GASLIMIT) // in units
	auth.GasPrice = gasPrice

	address, _, _, err := ERC20.DeployERC20Token(
		auth,
		client,
		big.NewInt(1000000000000), //10000亿枚，一枚等于1分钱
		"LianmiCoin",
		"LNMC",
	)
	if err != nil {
		log.Fatalf("deploy %v \n", err)
	}
	fmt.Printf("Contract pending deploy:  0x%x \n", address)

	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			fmt.Println(header.Hash().Hex()) // 0xf918192c4dd05834ebdee15920225f97d4ed350c829b007ae8ca95217f282f3e

			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(block.Number().Uint64())   // 1886
			fmt.Println(block.Time())              // 1601576368
			fmt.Println(block.Nonce())             // 3489525387448087149
			fmt.Println(len(block.Transactions())) // 1
		}
	}
}

//传参： 合约地址，账户地址
func getTokenBalance(contractAddress, accountAddress string) {
	client, err := ethclient.Dial(WSURI)
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	//使用合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(contractAddress), client)
	if err != nil {
		log.Fatalf("conn contract: %v \n", err)
	}

	//余额查询
	accountBalance, err := contract.BalanceOf(nil, common.HexToAddress(accountAddress))
	if err != nil {
		log.Fatalf("get Balances err: %v \n", err)
	}
	fmt.Println("Token of LNMC:", accountBalance)

}

//Eth转账, 从第0号叶子转到目标账号, amount  单位是 wei
func transferEth(sourcePrivateKey, target string, amount string) {
	client, err := ethclient.Dial(WSURI)
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
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

	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")
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

	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
}

//ERC20代币余额查询， 传参1是发送者合约地址，传参2是接收者账号地址
func querySendAndReceive(sender, receiver string) {

	client, err := ethclient.Dial(WSURI)
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
	client, err := ethclient.Dial(WSURI)
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

	client, err := ethclient.Dial(WSURI)
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

func main() {
	//部署合约于第1号叶子 ,  获取发币合约地址
	// deploy("fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915")
	//输出: Contract pending deploy: 0x1d2bdda8954b401feb52008c63878e698b6b8444

	//查询第1号叶子的LNMC余额
	getTokenBalance("0x1d2bdda8954b401feb52008c63878e698b6b8444", "0x4acea697f366C47757df8470e610a2d9B559DbBE")
	//输出: Token of LNMC: 10000000000

	// transfer("0x23a9497bb4ffa4b9d97d3288317c6495ecd3a2ce", "0xC74a1107faEEaB2994637902Ce4678432E262545", 400)
	//tx sent: 0x12139bdd617f66da7d123e20228e09092c5a55ebd2da9986c88fb1ec3cc55122

	// getTokenBalance("0x23a9497bb4ffa4b9d97d3288317c6495ecd3a2ce", "0x4acea697f366C47757df8470e610a2d9B559DbBE")
	//输出: Token of LNMC: 10000000000

	// getTokenBalance("0x23a9497bb4ffa4b9d97d3288317c6495ecd3a2ce", "0xC74a1107faEEaB2994637902Ce4678432E262545")

	// queryTx("0x40aa1ed6e2af939a9cc2f711a51cea0f21bdba3f146530f270956dbe3b454dd8")

	//查询Eth余额
	// queryETH("0xe14d151e0511b61357dde1b35a74e9c043c34c47")

}

/*
eth.sendTransaction({from:account1,to:account3,value:web3.toWei(1,"ether")})
*/
