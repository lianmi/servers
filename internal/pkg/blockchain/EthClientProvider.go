package blockchain

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"regexp"

	"github.com/lianmi/servers/internal/pkg/models"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/google/wire"
	"golang.org/x/crypto/sha3"
	// "github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/blockchain/hdwallet"
	ERC20 "github.com/lianmi/servers/internal/pkg/blockchain/lnmc/contracts/ERC20"
	MultiSig "github.com/lianmi/servers/internal/pkg/blockchain/lnmc/contracts/MultiSig"
	"github.com/lianmi/servers/internal/pkg/blockchain/util"
)

type KeyPair struct {
	PrivateKeyHex string //hex格式的私钥
	AddressHex    string //hex格式的地址
}

// Service service
type Service struct {
	o        *Options
	WsClient *ethclient.Client
	logger   *zap.Logger
}

// Options service options
type Options struct {
	WsURI                      string //websocket的 uri ws://127.0.0.1:8546
	ERC20DeployContractAddress string
}

func NewEthClientProviderOptions(v *viper.Viper, logger *zap.Logger) (*Options, error) {
	var err error
	o := new(Options)
	//读取dispatcher.yml配置文件里的eth设置
	if err = v.UnmarshalKey("ethereum", o); err != nil {
		return nil, errors.Wrap(err, "unmarshal ethereum option error")
	}
	wsUri := fmt.Sprintf("%s", o.WsURI)
	logger.Info("load ethereum options success", zap.String("WsUri", wsUri), zap.String("ERC20DeployContractAddress", o.ERC20DeployContractAddress))

	return o, err
}

// New returns new service
func New(opts *Options, logger *zap.Logger) (*Service, error) {
	if opts.WsURI == "" {
		return nil, errors.New("ethereum websocket uri is required")
	}
	client, err := ethclient.Dial("/etc/node/geth.ipc")
	if err != nil {
		logger.Error("ethclient.Dial ipc Failed", zap.String("IpcUri", "ipc:/etc/node/geth.ipc"), zap.Error(err))
		return nil, err
	} else {
		logger.Info("ethclient.Dial ipc succeed")
	}
	// client, err := ethclient.Dial(opts.WsURI)
	// if err != nil {
	// 	logger.Error("ethclient.Dial Failed", zap.String("WsUri", opts.WsURI), zap.Error(err))
	// 	return nil, err
	// }
	return &Service{
		o:        opts,
		WsClient: client,
		logger:   logger,
	}, nil
}

//关闭ws
func (s *Service) Stop() {
	s.WsClient.Close()
}

//创建系统HD钱包
func (s *Service) CreateHDWallet() {
	mnemonic := LMCommon.MnemonicServer // "element urban soda endless beach celery scheme wet envelope east glory retire"
	// fmt.Println("mnemonic:", mnemonic)

	wallet, err := hdwallet.NewFromMnemonic(mnemonic, LMCommon.SeedPassword)
	if err != nil {
		s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		return
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, true)
	if err != nil {
		s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		return
	}
	s.logger.Info("m/44'/60'/0'/0/0", zap.String("Account address", account.Address.Hex()))

	// privateKey, err := wallet.PrivateKey(account)
	// if err != nil {
	// 	s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
	// 	return
	// }
	// privateKeyHex, err := wallet.PrivateKeyHex(account)
	// if err != nil {
	// 	s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
	// 	return
	// }
	// fmt.Printf("Private key m/44'/60'/0'/0/0 in hex: %s\n", privateKeyHex)
	// s.logger.Info("m/44'/60'/0'/0/0", zap.String("Private key", privateKeyHex))

	publicKeyHex, _ := wallet.PublicKeyHex(account)
	if err != nil {
		s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		return
	}

	s.logger.Info("m/44'/60'/0'/0/0", zap.String("Public key", publicKeyHex))

	//第1号索引派生
	{
		path1 := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/1")
		account1, err := wallet.Derive(path1, true)
		if err != nil {
			s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
			return
		}

		// fmt.Printf("m/44'/60'/0'/0/1 Account address: %s\n", account1.Address.Hex())
		s.logger.Info("m/44'/60'/0'/0/1", zap.String("Account1 address", account1.Address.Hex()))

		// privateKey1, err := wallet.PrivateKey(account1)
		// if err != nil {
		// 	s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		// 	return
		// }
		// privateKeyHex1, err := wallet.PrivateKeyHex(account1)
		// if err != nil {
		// 	s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		// 	return
		// }
		// // fmt.Printf("m/44'/60'/0'/0/1 Private key in hex: %s\n", privateKeyHex1)
		// s.logger.Info("m/44'/60'/0'/0/1", zap.String("Private key", privateKeyHex1))

		publicKeyHex1, _ := wallet.PublicKeyHex(account1)
		if err != nil {
			s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
			return
		}

		s.logger.Info("m/44'/60'/0'/0/1", zap.String("Public key", publicKeyHex1))

	}

	//第2号索引派生
	{
		path2 := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/2")
		account2, err := wallet.Derive(path2, true)
		if err != nil {
			s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
			return
		}

		s.logger.Info("m/44'/60'/0'/0/2", zap.String("Account2 address", account2.Address.Hex()))

		// privateKey1, err := wallet.PrivateKey(account1)
		// if err != nil {
		// 	s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		// 	return
		// }
		// privateKeyHex1, err := wallet.PrivateKeyHex(account1)
		// if err != nil {
		// 	s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		// 	return
		// }
		// // fmt.Printf("m/44'/60'/0'/0/1 Private key in hex: %s\n", privateKeyHex1)
		// s.logger.Info("m/44'/60'/0'/0/1", zap.String("Private key", privateKeyHex1))

		publicKeyHex2, _ := wallet.PublicKeyHex(account2)
		if err != nil {
			s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
			return
		}

		s.logger.Info("m/44'/60'/0'/0/2", zap.String("Public key", publicKeyHex2))

	}
	/*
		2020-10-02T00:18:25.721+0800	INFO	m/44'/60'/0'/0/0	{"Account address": "0xe14D151e0511b61357DDe1B35a74E9c043c34C47"}
		2020-10-02T00:18:25.721+0800	INFO	m/44'/60'/0'/0/0	{"Private key": "4c88e6ccffec59b6c3df5ab51a4e6c42c421f58274d653d716aafd4aff376f5b"}
		2020-10-02T00:18:25.722+0800	INFO	m/44'/60'/0'/0/0	{"Public key": "b97cf13c8758594fb59c14765f365d05b9e67539e8f50721f8f6b8401f13af93e623ee620d9de8058b4043a0bc8be99e9135b6aa1c10e9ca8e85e0c4828e3070"}
		2020-10-02T00:18:25.722+0800	INFO	m/44'/60'/0'/0/1	{"Account1 address": "0x4acea697f366C47757df8470e610a2d9B559DbBE"}
		2020-10-02T00:18:25.723+0800	INFO	m/44'/60'/0'/0/1	{"Private key": "fb874fd86fc8e2e6ac0e3c2e3253606dfa10524296ee43d65f722965c5d57915"}
		2020-10-02T00:18:25.723+0800	INFO	m/44'/60'/0'/0/1	{"Public key": "553d2e5a5ad1ac9b2ae2dab3ddc28df74e1a549a753706715ec238e3e5c55008e45995b0d3271f8120890c74acc3602829207cefd432cfe1c1ca25767fd7a439"}
	*/
}

//根据叶子索引号获取到公私钥对
func (s *Service) GetKeyPairsFromLeafIndex(index uint64) *KeyPair {
	mnemonic := LMCommon.MnemonicServer // "element urban soda endless beach celery scheme wet envelope east glory retire"
	// fmt.Println("mnemonic:", mnemonic)

	wallet, err := hdwallet.NewFromMnemonic(mnemonic, LMCommon.SeedPassword)
	if err != nil {
		s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		return nil
	}
	leaf := fmt.Sprintf("m/44'/60'/0'/0/%d", index)
	path := hdwallet.MustParseDerivationPath(leaf)
	account, err := wallet.Derive(path, true)
	if err != nil {
		s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		return nil
	}
	// s.logger.Info(fmt.Sprintf("m/44'/60'/0'/0/%d", index), zap.String("Account address", account.Address.Hex()))

	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		return nil
	}
	privateKeyHex, err := wallet.PrivateKeyHex(account)
	if err != nil {
		s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		return nil
	}
	// fmt.Printf("Private key m/44'/60'/0'/0/0 in hex: %s\n", privateKeyHex)
	// s.logger.Info(fmt.Sprintf("m/44'/60'/0'/0/%d", index), zap.String("Private key", privateKeyHex))

	publicKeyHex, _ := wallet.PublicKeyHex(account)
	if err != nil {
		s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		return nil
	}

	// s.logger.Info(fmt.Sprintf("m/44'/60'/0'/0/%d", index), zap.String("Public key", publicKeyHex))

	_ = privateKey
	_ = publicKeyHex

	return &KeyPair{
		PrivateKeyHex: privateKeyHex,         //hex格式的私钥
		AddressHex:    account.Address.Hex(), //hex格式的地址
	}

}

// GetLatestBlockNumber get the latest block number
func (s *Service) GetLatestBlockNumber() (*big.Int, error) {

	block, err := s.WsClient.HeaderByNumber(context.Background(), nil)
	return block.Number, err
}

// GetPublicAddressFromPrivateKey returns public address from private key
func (s *Service) GetPublicAddressFromPrivateKey(priv *ecdsa.PrivateKey) (common.Address, error) {
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
func (s *Service) GetGasPrice() *big.Int {

	maxGasPrice := big.NewInt(9000000000)     // 9 gwei
	defaultGasPrice := big.NewInt(1000000000) // 1 gwei
	gasPrice, err := s.WsClient.SuggestGasPrice(context.Background())
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
func (s *Service) GetWeiBalance(address string) (uint64, error) {

	account := common.HexToAddress(address)
	balance, err := s.WsClient.BalanceAt(context.Background(), account, nil)
	if err != nil {
		// log.Fatal(err)
		s.logger.Error("BalanceAt ", zap.Error(err))
		return 0, err
	}
	// fmt.Println("balance: ", balance)
	return balance.Uint64(), nil

}

// 输出Eth为单位的账户余额
func (s *Service) GetEthBalance(address string) (float64, error) {

	account := common.HexToAddress(address)
	balance, err := s.WsClient.BalanceAt(context.Background(), account, nil)
	if err != nil {
		s.logger.Error("BalanceAt ", zap.Error(err))
		return 0, err
	}
	ethAmount := util.ToDecimal(balance, 18)
	f64, _ := ethAmount.Float64()
	// fmt.Println("balance(Eth): ", f64)
	return f64, nil
}

//CheckIsvalidAddress 返回地址(普通  地址) 是否合法, 无须链上查询
func (s *Service) CheckIsvalidAddress(addressHex string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")

	return re.MatchString(addressHex)

}

//CheckAddressIsvalidContract 返回地址(合约地址) 是否合法, 需要链上查找
func (s *Service) CheckAddressIsvalidContract(addressHex string) bool {

	address := common.HexToAddress(addressHex)
	bytecode, err := s.WsClient.CodeAt(context.Background(), address, nil) // nil is latest block
	if err != nil {
		s.logger.Error("CodeAt ", zap.Error(err))
		return false
	}

	isContract := len(bytecode) > 0
	return isContract
}

// 根据keystore文件与密码生成私钥
func (s *Service) KeystoreToPrivateKey(privateKeyFile, password string) (string, string, error) {
	keyjson, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		// fmt.Println("read keyjson file failed：", err)
		s.logger.Error("read keyjson file failed ", zap.Error(err))
		return "", "", err
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
func (s *Service) CheckTransactionReceipt(_txHash string) (*types.Receipt, error) {

	txHash := common.HexToHash(_txHash)
	receipt, err := s.WsClient.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

//订阅并检测交易是否成功, 并返回区块高度, 0- 表示失败
func (s *Service) WaitForBlockCompletation(hashToRead string) uint64 {
	headers := make(chan *types.Header)
	sub, err := s.WsClient.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		s.logger.Error("SubscribeNewHead failed ", zap.Error(err))
		return 0
	}

	for {
		select {
		case err := <-sub.Err():
			_ = err
			return 0
		case header := <-headers:
			// s.logger.Info(header.TxHash.Hex())
			receipt, err := s.CheckTransactionReceipt(hashToRead)
			if err != nil {
				s.logger.Error("CheckTransactionReceipt failed ", zap.Error(err))
			}
			transactionStatus := receipt.Status
			if transactionStatus == 0 {
				//FAILURE
				sub.Unsubscribe()
				return 0
			} else if transactionStatus == 1 {
				//SUCCESS
				block, err := s.WsClient.BlockByHash(context.Background(), header.Hash())
				if err != nil {
					s.logger.Error("BlockByHash failed ", zap.Error(err))
					return 0
				}
				// log.Println("区块: ", block.Hash().Hex())
				// log.Println("区块编号: ", block.Number().Uint64())
				s.logger.Info("区块信息", zap.String("Hash", block.Hash().Hex()), zap.Uint64("Number", block.Number().Uint64()))
				s.QueryTransactionByBlockNumber(block.Number().Uint64())
				sub.Unsubscribe()
				return block.Number().Uint64()
			}
		}
	}
}

//根据区块高度查询里面所有交易
func (s *Service) QueryTransactionByBlockNumber(number uint64) {

	blockNumber := big.NewInt(int64(number))
	block, err := s.WsClient.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		s.logger.Error("BlockByNumber failed ", zap.Error(err))
		return
	}
	// log.Println("=========queryTransactionByBlockNumber start==========")
	for _, tx := range block.Transactions() {
		gasCost := util.CalcGasCost(tx.Gas(), tx.GasPrice()) //计算交易所需要支付的总费用
		ethAmount := util.ToDecimal(gasCost, 18)
		ethAmountF64, _ := ethAmount.Float64()

		chainID, err := s.WsClient.NetworkID(context.Background())
		if err != nil {
			s.logger.Error("NetworkID failed ", zap.Error(err))

		}

		msg, err := tx.AsMessage(types.NewEIP155Signer(chainID))
		if err != nil {
			s.logger.Error("NetworkID failed ", zap.Error(err))
		} else {

		}

		receipt, err := s.WsClient.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			s.logger.Error("NetworkID failed ", zap.Error(err))
		}

		s.logger.Info("tx info",
			zap.String("Hash", tx.Hash().Hex()),
			zap.String("Value", tx.Value().String()),
			zap.Uint64("Gas", tx.Gas()),
			zap.Uint64("GasPrice", tx.GasPrice().Uint64()),
			zap.Uint64("交易总费用(Wei)", gasCost.Uint64()),
			zap.Float64("交易总费用(Eth)", ethAmountF64),
			zap.Uint64("Nonce", tx.Nonce()),
			zap.ByteString("Data", tx.Data()),
			// zap.String("To", tx.To().Hex()),
			zap.String("From", msg.From().Hex()),
			zap.Uint64("Status", receipt.Status), //1-succeed
		)

	}

	// log.Println("=========queryTransactionByBlockNumber end==========")
}

//根据交易查询里面详情
func (s *Service) QueryTxInfoByHash(txHashHex string) (*models.HashInfo, error) {
	txHash := common.HexToHash(txHashHex)

	tx, _, err := s.WsClient.TransactionByHash(context.Background(), txHash)
	if err != nil {
		s.logger.Error("QueryTxInfoByHash(), TransactionByHash failed ", zap.Error(err))
	}

	s.logger.Info("QueryTxInfoByHash",
		zap.String("Hash: ", tx.Hash().Hex()),
		zap.String("Value: ", tx.Value().String()),
		zap.Uint64("Gas: ", tx.Gas()),
		zap.String("Value: ", tx.Value().String()),
	)

	return &models.HashInfo{
		TxHash: tx.Hash().Hex(),
		Nonce:  tx.Nonce(),
		Gas:    tx.Gas(),
		Data:   hex.EncodeToString(tx.Data()),
	}, nil
}

//从第0号叶子账号地址转账Eth到其它普通账号地址, 以wei为单位, 1 eth = 1x18次方wei
//data是上链的数据
func (s *Service) TransferEthToOtherAccount(targetAccount string, amount int64, data []byte) (blockNumber uint64, hash string, err error) {

	//第0号叶子私钥
	privKeyHex := s.GetKeyPairsFromLeafIndex(LMCommon.ETHINDEX).PrivateKeyHex //使用0号叶子
	privateKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		s.logger.Error("BlockByNumber failed ", zap.Error(err))
		return 0, "", err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		s.logger.Error("cannot assert type: publicKey is not of type *ecdsa.PublicKe")
		return 0, "", errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKe")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := s.WsClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		s.logger.Error("PendingNonceAt failed ", zap.Error(err))
		return 0, "", err
	}

	value := big.NewInt(amount)           // in wei (1 eth)
	gasLimit := uint64(LMCommon.GASLIMIT) // in units
	gasPrice := s.GetGasPrice()

	//接收账号
	toAddress := common.HexToAddress(targetAccount)

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := s.WsClient.NetworkID(context.Background())
	if err != nil {
		s.logger.Error("NetworkID failed ", zap.Error(err))
		return 0, "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		s.logger.Error("SignTx failed ", zap.Error(err))
		return 0, "", err
	}

	err = s.WsClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		s.logger.Error("TransferEthToOtherAccount(), SendTransaction failed ", zap.Error(err))
		return 0, "", err
	}

	s.logger.Info("tx sent", zap.String("Hash", signedTx.Hash().Hex()))

	/*
		等待检测交易是否完成，挖矿工需要工作才能出块
		> miner.start()
		> var account2="0x4acea697f366C47757df8470e610a2d9B559DbBE"
		> web3.fromWei(web3.eth.getBalance(account2), 'ether')
		输出： 1
	*/

	blockNumber = s.WaitForBlockCompletation(signedTx.Hash().Hex())
	if blockNumber > 0 {

		tx, isPending, err := s.WsClient.TransactionByHash(context.Background(), signedTx.Hash())
		if err != nil {
			s.logger.Error("TransferEthToOtherAccount(), TransactionByHash failed ", zap.Error(err))
		}

		s.logger.Info("交易完成",
			zap.Uint64("区块高度: ", blockNumber),
			zap.String("交易哈希: ", tx.Hash().Hex()),
			zap.Bool("isPending: ", isPending),
		)
		return blockNumber, tx.Hash().Hex(), nil

	} else {
		s.logger.Error("交易失败")
		return 0, "", errors.New("交易失败")
	}

}

//获取 LNMC余额, 以分为单位，无小数点, 传参： 账户地址
func (s *Service) GetLNMCTokenBalance(accountAddress string) (uint64, error) {

	//使用合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(s.o.ERC20DeployContractAddress), s.WsClient)
	if err != nil {
		s.logger.Error("conn contracts failed ", zap.Error(err))
		return 0, err
	}

	//余额查询
	accountLNMCBalance, err := contract.BalanceOf(nil, common.HexToAddress(accountAddress))
	if err != nil {
		s.logger.Error("get LNMC Balances failed ", zap.Error(err))
		return 0, err
	}
	s.logger.Debug("Token of LNMC:", zap.String("accountAddress", accountAddress), zap.String("Balance", accountLNMCBalance.String()))
	return accountLNMCBalance.Uint64(), nil

}

//从ERC20代币总账号转账到目标普通账号
func (s *Service) TransferLNMCFromLeaf1ToNormalAddress(target string, amount int64) (blockNumber uint64, hash string, amountAfter uint64, err error) {
	var privateKeyHex string
	privateKeyHex = s.GetKeyPairsFromLeafIndex(LMCommon.LNMCINDEX).PrivateKeyHex //使用1号叶子

	//查询转账之前的LNMC余额
	amountCurrent, err := s.GetLNMCTokenBalance(target)
	if err != nil {
		s.logger.Error("GetLNMCTokenBalance failed ", zap.Error(err))
		return 0, "", 0, err
	} else {
		s.logger.Info("查询转账之前的LNMC余额", zap.Uint64("amountCurrent", amountCurrent))
	}

	//使用第1号叶子的发币合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(s.o.ERC20DeployContractAddress), s.WsClient)
	if err != nil {
		// log.Fatalf("conn contract: %v \n", err)
		s.logger.Error("conn contracts failed ", zap.Error(err))
		return 0, "", 0, err
	}
	// 第1号叶子私钥
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		s.logger.Error("HexToECDSA failed ", zap.Error(err))
		return 0, "", 0, err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		s.logger.Error("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return 0, "", 0, errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := s.WsClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		s.logger.Error("PendingNonceAt failed ", zap.Error(err))
		return 0, "", 0, err
	}

	gasPrice, err := s.WsClient.SuggestGasPrice(context.Background())
	if err != nil {
		s.logger.Error("SuggestGasPrice failed ", zap.Error(err))
		return 0, "", 0, err
	}

	auth := bind.NewKeyedTransactor(privateKey) //第1号叶子的子私钥
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)                // in wei
	auth.GasLimit = uint64(LMCommon.GASLIMIT) // in units
	auth.GasPrice = gasPrice

	//调用合约里的转账函数
	contractTx, err := contract.Transfer(&bind.TransactOpts{
		From:   auth.From,
		Signer: auth.Signer,
		Value:  nil,
	}, common.HexToAddress(target), big.NewInt(amount))
	if err != nil {
		// log.Fatalf("TransferFrom err: %v \n", err)
		s.logger.Error("TransferLNMCFromLeaf1ToNormalAddress, TransferFrom failed", zap.Error(err))
		return 0, "", 0, err
	}
	s.logger.Info("tx sent", zap.String("Hash", contractTx.Hash().Hex()))

	//监听交易直到打包完成
	blockNumber = s.WaitForBlockCompletation(contractTx.Hash().Hex())
	if blockNumber > 0 {

		tx, isPending, err := s.WsClient.TransactionByHash(context.Background(), contractTx.Hash())
		if err != nil {
			s.logger.Error("TransactionByHash failed ", zap.Error(err))
		}

		//查询转账之后的LNMC余额
		amountAfter, err := s.GetLNMCTokenBalance(target)
		if err != nil {
			s.logger.Error("GetLNMCTokenBalance failed ", zap.Error(err))
		}

		s.logger.Info("交易完成",
			zap.String("交易哈希: ", tx.Hash().Hex()),
			zap.Bool("isPending: ", isPending),
			zap.Uint64("amountAfter: ", amountAfter),
		)
		return blockNumber, tx.Hash().Hex(), amountAfter, nil

	} else {
		s.logger.Error("交易失败")
		return 0, "", 0, errors.New("交易失败")
	}
}

/*
部署多签合约
约定： 使用1号叶子地址作为发币的私钥
*/
func (s *Service) DeployMultiSig(addressHexA, addressHexB string) (contractAddress string, blockNumber uint64, hash string, err error) {

	privateKeyHex := s.GetKeyPairsFromLeafIndex(LMCommon.LNMCINDEX).PrivateKeyHex //使用1号叶子
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		s.logger.Error("HexToECDSA failed ", zap.Error(err))
		return "", 0, "", err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		s.logger.Error("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return "", 0, "", errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := s.WsClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		s.logger.Error("PendingNonceAt failed ", zap.Error(err))
		return "", 0, "", err
	}

	gasPrice, err := s.WsClient.SuggestGasPrice(context.Background())
	if err != nil {
		s.logger.Error("SuggestGasPrice failed ", zap.Error(err))
		return "", 0, "", err
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)                // in wei
	auth.GasLimit = uint64(LMCommon.GASLIMIT) // in units
	auth.GasPrice = gasPrice

	address, deployMultiSigTx, _, err := MultiSig.DeployMultiSig(
		auth,
		s.WsClient,
		common.HexToAddress(addressHexA), //A 账号地址  发起者
		common.HexToAddress(addressHexB), //B 账号地址  证明人
		common.HexToAddress(s.o.ERC20DeployContractAddress), //ERC20发币地址
	)
	if err != nil {
		s.logger.Error("DeployMultiSig failed ", zap.Error(err))
		return "", 0, "", err
	}

	contractAddress = address.String()

	s.logger.Info("Contract pending deploy succeed",
		zap.String("deploy", address.String()),
		zap.String("Hash", deployMultiSigTx.Hash().String()),
	)

	//监听，直到合约部署成功,如果失败，则提示
	blockNumber = s.WaitForBlockCompletation(deployMultiSigTx.Hash().Hex())
	if blockNumber > 0 {

		tx, isPending, err := s.WsClient.TransactionByHash(context.Background(), deployMultiSigTx.Hash())
		if err != nil {
			s.logger.Error("TransactionByHash failed ", zap.Error(err))
			return "", 0, "", err
		}
		s.logger.Info("多签合约部署成功",
			zap.String("contractAddress", contractAddress),
			zap.String("Hash", tx.Hash().Hex()),
			zap.Bool("isPending", isPending),
		)
		return contractAddress, blockNumber, tx.Hash().Hex(), nil

	} else {
		s.logger.Error("多签合约部署失败")
		return "", 0, "", errors.New("多签合约部署失败")
	}

}

/*
代币转账到多签合约账户, sourcePrivateKey是发起转账的用户
发币合约地址由walletservice.yml指定
一般用于发币账户转到 多签智能合约及其它账号
如果是充值，则第一个参数是第1号叶子的私钥
*/
func (s *Service) TransferLNMCTokenToAddress(sourcePrivateKey, target string, amount int64) (uint64, string, error) {

	//使用总发币的合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(s.o.ERC20DeployContractAddress), s.WsClient)
	if err != nil {
		s.logger.Error("NewERC20Token failed ", zap.Error(err))
		return 0, "", err
	}

	//发起转账的用户的私钥
	privateKey, err := crypto.HexToECDSA(sourcePrivateKey)
	if err != nil {
		s.logger.Error("HexToECDSA failed ", zap.Error(err))
		return 0, "", err
	}
	auth := bind.NewKeyedTransactor(privateKey)

	//调用合约里的转账函数
	transferTx, err := contract.Transfer(&bind.TransactOpts{
		From:   auth.From,   //发起转账的用户账户地址
		Signer: auth.Signer, //签名
		Value:  nil,
	}, common.HexToAddress(target), big.NewInt(amount))
	if err != nil {

		s.logger.Error("TransferLNMCTokenToAddress, TransferFrom failed", zap.Error(err))
		return 0, "", err
	}
	// fmt.Printf("tx sent: %s \n", transferTx.Hash().Hex())

	//监听，直到转账成功,如果失败，则提示
	blockNumber := s.WaitForBlockCompletation(transferTx.Hash().Hex())
	if blockNumber > 0 {
		tx, isPending, err := s.WsClient.TransactionByHash(context.Background(), transferTx.Hash())
		if err != nil {
			s.logger.Error("TransactionByHash, TransferFrom failed", zap.Error(err))
			return 0, "", err
		}
		s.logger.Info("转账成功",
			zap.String("Hash", tx.Hash().Hex()),
			zap.Bool("isPending", isPending),
		)

		return blockNumber, tx.Hash().Hex(), nil
	} else {
		s.logger.Error("代币转账到目标账户失败")
		return 0, "", errors.New("代币转账到目标账户失败")
	}

}

/*
构造一个普通用户账号转账的裸交易数据,  用于预支付的发起
source - 发起方钱包账号
target - 接收者的钱包地址
tokens - 代币数量，字符串格式
*/
func (s *Service) GenerateTransferLNMCTokenTx(redisConn redis.Conn, source, target string, tokens int64) (*models.RawDesc, error) {
	var err error
	var balanceEth uint64 //用户当前ETH数量

	s.logger.Debug("GenerateTransferLNMCTokenTx start...",
		zap.String("source", source),
		zap.String("target", target),
		zap.Int64("tokens", tokens),
	)

	//当前用户的链上Eth余额
	balanceEth, err = s.GetWeiBalance(source)
	if err != nil {
		return nil, err
	}
	if balanceEth < LMCommon.GASLIMIT {
		s.logger.Error("用户链上Eth余额不足以支付交易gas手续费 ",
			zap.String("walletAddress", source),
			zap.Uint64("当前余额 balanceEth", balanceEth),
		)
		return nil, errors.New("Not sufficient funds")
	}
	fromAddress := common.HexToAddress(source)

	successNonceAt, err := s.WsClient.NonceAt(context.Background(), fromAddress, nil)
	if err != nil {
		s.logger.Error("Get NonceAt failed ", zap.Error(err))
		return nil, err
	}

	nonce, err := s.WsClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		s.logger.Error("PendingNonceAt failed", zap.Error(err))
		return nil, err
	}
	//TODO 从redis里取出上次的 PendingNonceAt 假如这两个nonce都相同，那么报错
	nonceAtKey := fmt.Sprintf("PendingNonceAt:%s", source)
	oldPendingNonceAt, err := redis.Uint64(redisConn.Do("GET", nonceAtKey))

	if oldPendingNonceAt == nonce {
		s.logger.Error("oldPendingNonceAt 等于 nonce, 不能上链交易")
		// return nil, errors.Wrapf(err, "oldPendingNonceAt 等于 nonce, 不能上链交易")
	}

	//TODO 这里有幺蛾子，nonce不会增长,连续发起多个交易会堵塞
	// see : https://blog.csdn.net/sinat_34070003/article/details/79919431
	// see: https://blog.csdn.net/qq_44373419/article/details/106492988 golang 实现 ETH 交易离线签名（冷签）--以太坊DPOS

	s.logger.Debug("Get NonceAt succeed",
		zap.Uint64("successNonceAt", successNonceAt),
		zap.Uint64("PendingNonceAt", nonce),
		zap.Uint64("nonce的值相差", nonce-successNonceAt),
	)
	_, err = redisConn.Do("SET", nonceAtKey, nonce)

	// s.logger.Debug("Generate TransferLNMCTokenTx succeed", zap.Int64("nonce", int64(nonce)), zap.Int64("tokens", tokens))

	value := big.NewInt(0) // in wei (0 eth) 由于进行的是代币转账，不设计以太币转账，因此这里填0
	gasPrice, err := s.WsClient.SuggestGasPrice(context.Background())
	if err != nil {
		s.logger.Error("SuggestGasPrice failed ", zap.Error(err))
		return nil, err
	}

	//接收者的钱包地址
	toAddress := common.HexToAddress(target)

	//注意，这里需要填写发币合约地址
	tokenAddress := common.HexToAddress(s.o.ERC20DeployContractAddress)

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	// fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	// fmt.Println(hexutil.Encode(paddedAddress))

	amount := big.NewInt(tokens) //代币数量

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	gasLimit := uint64(LMCommon.GASLIMIT) //必须强行指定，否则无法打包

	//构造代币转账的交易裸数据
	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)

	chainID, err := s.WsClient.NetworkID(context.Background())
	if err != nil {
		s.logger.Error("NetworkID failed ", zap.Error(err))
		return nil, err
	}
	// fmt.Println("chainID:", chainID.String())

	return &models.RawDesc{
		Nonce:           nonce,
		GasPrice:        gasPrice.Uint64(),
		GasLimit:        gasLimit,
		ChainID:         chainID.Uint64(),
		Txdata:          data,
		Value:           0,
		TxHash:          tx.Hash().Hex(), //已经生成的
		ContractAddress: s.o.ERC20DeployContractAddress,
	}, nil
}

//ERC20代币余额查询， 传参: 账号地址
func (s *Service) QueryLNMCBalance(addressHex string) (int64, error) {

	//使用发币合约地址
	contract, err := ERC20.NewERC20Token(common.HexToAddress(s.o.ERC20DeployContractAddress), s.WsClient)
	if err != nil {
		s.logger.Error("NewERC20Token failed ", zap.Error(err))
		return 0, err
	}

	if accountBalance, err := contract.BalanceOf(nil, common.HexToAddress(addressHex)); err != nil {
		s.logger.Error("BalanceOf failed ", zap.Error(err))
		return 0, err
	} else {

		// fmt.Printf("账号[%s]的LNMC余额: %s LNMC\n", addressHex, accountBalance.String())
		return accountBalance.Int64(), nil
	}

}

//根据客户端SDK签名后的裸交易数据，广播到链上
func (s *Service) SendSignedTxToGeth(rawTxHex string) (uint64, string, error) {
	var blockHash string
	rawTxBytes, err := hex.DecodeString(rawTxHex)

	signedTx := new(types.Transaction)
	rlp.DecodeBytes(rawTxBytes, &signedTx)

	err = s.WsClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		s.logger.Error("SendSignedTxToGeth(), SendTransaction failed ", zap.Error(err))
		return 0, "", err
	}

	// fmt.Printf("signedTx sent: %s", signedTx.Hash().Hex())

	//等待打包完成的回调
	blockNumber := s.WaitForBlockCompletation(signedTx.Hash().Hex())
	if blockNumber > 0 {
		//获取交易哈希里的打包状态，如果打包完成，isPending = false
		tx2, isPending, err := s.WsClient.TransactionByHash(context.Background(), signedTx.Hash())
		if err != nil {
			s.logger.Error("TransactionByHash failed ", zap.Error(err))
			return 0, "", err
		}

		blockHash = tx2.Hash().Hex()
		s.logger.Info("SendSignedTxToGeth",
			zap.String("Hash", tx2.Hash().Hex()),
			zap.Bool("isPending", isPending),
		)

	} else {
		// log.Println(" 打包失败")
		s.logger.Error("SendSignedTxToGeth失败")
		return 0, "", errors.New("SendSignedTxToGeth failed")
	}
	return blockNumber, blockHash, nil
}

/*
多签合约, 从合约账号转到目标账号
传参：
  1. multiSigContractAddress 第一步部署的多签智能合约， A+B => C
  2. privateKeySource A或B的私钥，用来签名
  3. target 目标地址C, 在本系统里，需要派生一个子地址来接收

*/
func (s *Service) TransferTokenFromABToC(multiSigContractAddress, privateKeySource, target string, amount int64) error {

	cAddr := common.HexToAddress(multiSigContractAddress)
	// fmt.Println(cAddr.String())

	//调用多签智能合约地址
	contract, err := MultiSig.NewMultiSig(cAddr, s.WsClient)
	if err != nil {
		s.logger.Error("NewMultiSig failed ", zap.Error(err))
		return err
	}
	fmt.Println(contract.Name(&bind.CallOpts{Pending: true}))

	privateKey, err := crypto.HexToECDSA(privateKeySource) //A或B私钥
	if err != nil {
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		// log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		s.logger.Error("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")

	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := s.WsClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		s.logger.Error("PendingNonceAt failed ", zap.Error(err))
		return err
	}

	gasPrice, err := s.WsClient.SuggestGasPrice(context.Background())
	if err != nil {
		s.logger.Error("SuggestGasPrice failed ", zap.Error(err))
		return err
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)                // in wei
	auth.GasLimit = uint64(LMCommon.GASLIMIT) // in units
	auth.GasPrice = gasPrice

	//调用合约里的转账函数
	transferMultiSigTx, err := contract.Transfer(&bind.TransactOpts{
		From:   auth.From,
		Signer: auth.Signer,
		Value:  nil,
	}, common.HexToAddress(target), big.NewInt(amount)) //LNMC
	if err != nil {

		s.logger.Error("Transfer failed ", zap.Error(err))
		return err
	}
	// fmt.Printf("tx of multisig contract sent: %s \n", transferMultiSigTx.Hash().Hex())

	done := s.WaitForBlockCompletation(transferMultiSigTx.Hash().Hex())
	if done == 1 {
		tx2, isPending, err := s.WsClient.TransactionByHash(context.Background(), transferMultiSigTx.Hash())
		if err != nil {

			s.logger.Error("TransactionByHash failed", zap.Error(err))
			return err
		}

		s.logger.Info("multisig contract 打包成功",
			zap.String("Hash", tx2.Hash().Hex()),
			zap.Bool("isPending", isPending),
		)
		return nil

	} else {
		s.logger.Error("multisig contract 打包失败")
		return errors.New("multisig contract 打包失败")
	}

}

//根据多签合约生成裸交易数据
func (s *Service) GenerateRawTx(contractAddress, fromAddressHex, target string, tokens int64) (*models.RawDesc, error) {
	fromAddress := common.HexToAddress(fromAddressHex)
	nonce, err := s.WsClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		s.logger.Error("PendingNonceAt failed", zap.Error(err))
		return nil, err
	}
	// fmt.Println("nonce:", int64(nonce))
	nonce = nonce + 1
	value := big.NewInt(0) // in wei (0 eth) 由于进行的是代币转账，不设计以太币转账，因此这里填0
	gasPrice, err := s.WsClient.SuggestGasPrice(context.Background())
	if err != nil {
		s.logger.Error("SuggestGasPrice failed", zap.Error(err))
		return nil, err
	}

	// fmt.Println("gasPrice", gasPrice)

	//接收者地址：用户D
	toAddress := common.HexToAddress(target)

	//注意，这里需要填写多签合约地址
	tokenAddress := common.HexToAddress(contractAddress)

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	// fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	// fmt.Println(hexutil.Encode(paddedAddress)) // 0x00000000000000000000000059ac768b416c035c8db50b4f54faaa1e423c070d

	amount := big.NewInt(tokens) //代币数量

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	// fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	// fmt.Println("data:", data)
	// fmt.Println("data hex:", hex.EncodeToString(data))

	gasLimit := uint64(LMCommon.GASLIMIT) //必须强行指定，否则无法打包
	// fmt.Println("gasLimit:", gasLimit)

	//构造代币转账的交易裸数据
	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)

	chainID, err := s.WsClient.NetworkID(context.Background())
	if err != nil {
		s.logger.Error("NetworkID failed", zap.Error(err))
		return nil, err
	}
	// fmt.Println("chainID:", chainID.String())

	// rawData, _ := tx.MarshalJSON()

	// _ = tx

	return &models.RawDesc{
		Nonce:           nonce,
		GasPrice:        gasPrice.Uint64(),
		GasLimit:        gasLimit,
		ChainID:         chainID.Uint64(),
		Txdata:          data,
		Value:           0,
		TxHash:          tx.Hash().Hex(),
		ContractAddress: s.o.ERC20DeployContractAddress,
	}, nil

}

var ProviderSet = wire.NewSet(New, NewEthClientProviderOptions)
