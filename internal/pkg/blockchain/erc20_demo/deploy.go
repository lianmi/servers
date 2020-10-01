package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/lianmi/servers/internal/pkg/blockchain/erc20_demo/contracts"
	// "io/ioutil"
	"log"
	"math/big"
	// "strings"
)

/*
发币合约部署
第1号叶子:
privateKey: fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915
address: 0x4acea697f366C47757df8470e610a2d9B559DbBE
*/

func deploy(privateKeyHex string) {
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

	address, _, _, err := contracts.DeployERC20Token(
		auth,
		client,
		big.NewInt(10000000000), //100亿
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
	client, err := ethclient.Dial("ws://127.0.0.1:8546")
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	//使用合约地址
	contract, err := contracts.NewERC20Token(common.HexToAddress(contractAddress), client)
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

//ERC20代币转账, 从合约账号转到目标账号
func transfer(contractAddress, target string, amount int64) {
	/*
		blockchain, err := ethclient.Dial("http://127.0.0.1:8545")
		if err != nil {
			log.Fatalf("Unable to connect to network:%v \n", err)
		}

		//使用合约地址
		contract, err := contracts.NewERC20Token(common.HexToAddress(contractAddress), blockchain)
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
	*/
}

//ERC20代币余额查询， 传参1是发送者合约地址，传参2是接收者账号地址
func querySendAndReceive(sender, receiver string) {
	/*
		blockchain, err := ethclient.Dial("http://127.0.0.1:8545")
		if err != nil {
			log.Fatalf("Unable to connect to network:%v \n", err)
		}

		//使用合约地址
		contract, err := contracts.NewERC20Token(common.HexToAddress(sender), blockchain)
		if err != nil {
			log.Fatalf("conn contract: %v \n", err)
		}
		data, _ := ioutil.ReadFile(key)
		auth, err := bind.NewTransactor(strings.NewReader(string(data)), "123456")
		if err != nil {
			log.Fatalf("Failed to create authorized transactor:%v \n", err)
		}

		var accountBalance = big.NewInt(0)
		if accountBalance, err = contract.BalanceOf(nil, auth.From); err != nil {
			log.Fatalf("get Balances err: %v \n", err)
		}
		fmt.Println("发送方余额: ", accountBalance)

		// target是接收地址
		if accountBalance, err = contract.BalanceOf(nil, common.HexToAddress(receiver)); err != nil {
			log.Fatalf("get Balances err: %v \n", err)
		}
		fmt.Println("接收方余额: ", accountBalance)
	*/
}

func queryTx(txHex string) {
	txHash := common.HexToHash(txHex)
	blockchain, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	tx, isPending, err := blockchain.TransactionByHash(context.Background(), txHash)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(tx.Hash().Hex()) // 0x40aa1ed6e2af939a9cc2f711a51cea0f21bdba3f146530f270956dbe3b454dd8
	fmt.Println(isPending)       // false
}

//查询Eth余额
func queryETH(accountHex string) {

	blockchain, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Unable to connect to network:%v \n", err)
	}

	account := common.HexToAddress(accountHex)
	balance, err := blockchain.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(accountHex, ":", balance) // 1000000000000000000

}

func main() {
	//部署合约
	// deploy("fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915")
	//输出: Contract pending deploy: 0x23a9497bb4ffa4b9d97d3288317c6495ecd3a2ce

	//查询第1号叶子的LNMC余额
	getTokenBalance("0x23a9497bb4ffa4b9d97d3288317c6495ecd3a2ce", "0x4acea697f366C47757df8470e610a2d9B559DbBE")
	//输出: Token of LNMC: 10000000000

	// transfer("0xdeb284d75f757ce5e3c5de349732c05baa53584f", "0x8159077856ca85b20479f6ac6694ffed1d27fdf3", 400)
	//tx sent: 0x12139bdd617f66da7d123e20228e09092c5a55ebd2da9986c88fb1ec3cc55122
	// 发送方余额:  9999998960
	// 接收方余额:  1040

	//向HD钱包第0号索引派生的账号：  0xe14d151e0511b61357dde1b35a74e9c043c34c47 转账eth
	//向HD钱包第1号索引派生的账号：  0x4acea697f366C47757df8470e610a2d9B559DbBE 转账LNMC
	// transfer("0xc5597646fe9a9fd5057fff15df28af4ac78e992e", "0x4acea697f366C47757df8470e610a2d9B559DbBE", 200)
	//输出： 0x40aa1ed6e2af939a9cc2f711a51cea0f21bdba3f146530f270956dbe3b454dd8

	//查询 Tx
	// queryTx("0x40aa1ed6e2af939a9cc2f711a51cea0f21bdba3f146530f270956dbe3b454dd8")

	// 查询ERC20余额
	// querySendAndReceive("0xdeb284d75f757ce5e3c5de349732c05baa53584f", "0x8159077856ca85b20479f6ac6694ffed1d27fdf3")
	//
	//查询Eth余额
	// queryETH("0xe14d151e0511b61357dde1b35a74e9c043c34c47")

}

/*
eth.sendTransaction({from:account1,to:account3,value:web3.toWei(1,"ether")})
*/
