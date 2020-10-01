package blockchain

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/pkg/blockchain/util"
	"github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"io/ioutil"
	"math/big"
	// "regexp"
)

const (
	// KEY      = "UTC--2020-09-29T09-27-24.693765000Z--b18db89641d2ec807104258e2205e6ac6264bf25"

	// PASSWORD = "LianmiSky8900388"
	GASLIMIT = 3000000 //30000000000
)

// Service service
type Service struct {
	WsClient *ethclient.Client
	logger   *zap.Logger
}

// Options service options
type Options struct {
	WsURI string //websocket的 uri ws://127.0.0.1:8546
}

func NewEthClientProviderOptions(v *viper.Viper, logger *zap.Logger) (*Options, error) {
	var err error
	o := new(Options)
	//读取dispatcher.yaml配置文件里的redis设置
	if err = v.UnmarshalKey("ethereum", o); err != nil {
		return nil, errors.Wrap(err, "unmarshal ethereum option error")
	}
	wsUri := fmt.Sprintf("%s", o.WsURI)
	logger.Info("load ethereum options success", zap.String("WsUri", wsUri))

	return o, err
}

// New returns new service
func New(opts *Options, logger *zap.Logger) (*Service, error) {
	if opts.WsURI == "" {
		return nil, errors.New("ethereum websocket uri is required")
	}
	client, err := ethclient.Dial(opts.WsURI)
	if err != nil {
		return nil, err
	}
	return &Service{
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
	// mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"
	mnemonic := "element urban soda endless beach celery scheme wet envelope east glory retire"
	// fmt.Println("mnemonic:", mnemonic)

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
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

	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		return
	}
	privateKeyHex, err := wallet.PrivateKeyHex(account)
	if err != nil {
		s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		return
	}
	// fmt.Printf("Private key m/44'/60'/0'/0/0 in hex: %s\n", privateKeyHex)
	s.logger.Info("m/44'/60'/0'/0/0", zap.String("Private key", privateKeyHex))

	publicKeyHex, _ := wallet.PublicKeyHex(account)
	if err != nil {
		s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
		return
	}

	s.logger.Info("m/44'/60'/0'/0/0", zap.String("Public key", publicKeyHex))

	_ = privateKey

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

		privateKey1, err := wallet.PrivateKey(account1)
		if err != nil {
			s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
			return
		}
		privateKeyHex1, err := wallet.PrivateKeyHex(account1)
		if err != nil {
			s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
			return
		}
		// fmt.Printf("m/44'/60'/0'/0/1 Private key in hex: %s\n", privateKeyHex1)
		s.logger.Info("m/44'/60'/0'/0/1", zap.String("Private key", privateKeyHex1))

		publicKeyHex1, _ := wallet.PublicKeyHex(account1)
		if err != nil {
			s.logger.Error("NewFromMnemonic error", zap.String("err", err.Error()))
			return
		}

		s.logger.Info("m/44'/60'/0'/0/1", zap.String("Public key", publicKeyHex1))

		_ = privateKey1
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
func (s *Service) GetWeiBalance(address string) *big.Int {

	account := common.HexToAddress(address)
	balance, err := s.WsClient.BalanceAt(context.Background(), account, nil)
	if err != nil {
		// log.Fatal(err)
		s.logger.Error("BalanceAt ", zap.Error(err))
		return nil
	}
	// fmt.Println("balance: ", balance)
	return balance

}

// 输出Eth为单位的账户余额
func (s *Service) GetEthBalance(address string) float64 {

	account := common.HexToAddress(address)
	balance, err := s.WsClient.BalanceAt(context.Background(), account, nil)
	if err != nil {
		s.logger.Error("BalanceAt ", zap.Error(err))
		return 0
	}
	ethAmount := util.ToDecimal(balance, 18)
	f64, _ := ethAmount.Float64()
	// fmt.Println("balance(Eth): ", f64)
	return f64
}

//CheckAddressIsvalid 返回地址(普通地址或合约地址) 是否合法
func (s *Service) CheckAddressIsvalid(addressHex string) bool {

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
func (s *Service) CheckTransactionReceipt(_txHash string) int {

	txHash := common.HexToHash(_txHash)
	receipt, err := s.WsClient.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return (-1)
	}
	return (int(receipt.Status))
}

//订阅并检测交易是否成功
func (s *Service) WaitForBlockCompletation(hashToRead string) int {
	headers := make(chan *types.Header)
	sub, err := s.WsClient.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		s.logger.Error("SubscribeNewHead failed ", zap.Error(err))
		return -1
	}

	for {
		select {
		case err := <-sub.Err():
			_ = err
			return -1
		case header := <-headers:
			s.logger.Info(header.TxHash.Hex())
			transactionStatus := s.CheckTransactionReceipt(hashToRead)
			if transactionStatus == 0 {
				//FAILURE
				sub.Unsubscribe()
				return 0
			} else if transactionStatus == 1 {
				//SUCCESS
				block, err := s.WsClient.BlockByHash(context.Background(), header.Hash())
				if err != nil {
					s.logger.Error("BlockByHash failed ", zap.Error(err))
					return -1
				}
				// log.Println("区块: ", block.Hash().Hex())
				// log.Println("区块编号: ", block.Number().Uint64())
				s.logger.Info("区块信息", zap.String("Hash", block.Hash().Hex()), zap.Uint64("Number", block.Number().Uint64()))
				s.QueryTransactionByBlockNumber(block.Number().Uint64())
				sub.Unsubscribe()
				return 1
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
		// log.Println("tx.Hash: ", tx.Hash().Hex())            //
		// log.Println("tx.Value: ", tx.Value().String())       // 10000000000000000
		// log.Println("tx.Gas: ", tx.Gas())                    // 3000000
		// log.Println("tx.GasPrice: ", tx.GasPrice().Uint64()) // 1000000000

		// cost := tx.Gas() * tx.GasPrice().Uint64() //计算交易所需要支付的总费用
		gasCost := util.CalcGasCost(tx.Gas(), tx.GasPrice()) //计算交易所需要支付的总费用
		// log.Println("交易总费用(Wei): ", gasCost)                 //3000000000000000Wei
		ethAmount := util.ToDecimal(gasCost, 18)
		ethAmountF64, _ := ethAmount.Float64()
		// log.Println("交易总费用(eth): ", ethAmount) //0.003Eth
		// log.Println("tx.Nonce: ", tx.Nonce())
		// log.Println("tx.Data: ", tx.Data())
		// log.Println("tx.To: ", tx.To().Hex()) //目标地址

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
			zap.String("To", tx.To().Hex()),
			zap.String("From", msg.From().Hex()),
			zap.Uint64("Status", receipt.Status),
		)

		// log.Println(receipt.Status) // 1

	}

	// log.Println("=========queryTransactionByBlockNumber end==========")
}

//从总账号地址转账Eth到其它普通账号地址, 以wei为单位, 1 eth = 1x18次方wei
func (s *Service) TransferEthFromCoinbaseToOtherAccount(targetAccount string, amount int64) error {

	//主账号私钥
	privKeyHex := "bc2f812f1f534c9e8a3b3cfb628b0ea5d41967d4f18391c6489737d743b1ee7a"

	privateKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		s.logger.Error("BlockByNumber failed ", zap.Error(err))
		return err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		s.logger.Error("cannot assert type: publicKey is not of type *ecdsa.PublicKe")
		return errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKe")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := s.WsClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		s.logger.Error("PendingNonceAt failed ", zap.Error(err))
		return err
	}

	value := big.NewInt(amount)  // in wei (1 eth)
	gasLimit := uint64(GASLIMIT) // in units
	/*
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			log.Fatal(err)
		} else {
			log.Println("gasPrice: ", gasPrice)
		}
	*/
	gasPrice := s.GetGasPrice()

	//接收账号
	toAddress := common.HexToAddress(targetAccount)

	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := s.WsClient.NetworkID(context.Background())
	if err != nil {
		s.logger.Error("NetworkID failed ", zap.Error(err))
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		s.logger.Error("SignTx failed ", zap.Error(err))
		return err
	}

	err = s.WsClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		s.logger.Error("SendTransaction failed ", zap.Error(err))
		return err
	}

	s.logger.Info("tx sent", zap.String("Hash", signedTx.Hash().Hex()))

	/*
		等待检测交易是否完成，挖矿工需要工作才能出块
		> miner.start()
		> var account2="0x4acea697f366C47757df8470e610a2d9B559DbBE"
		> web3.fromWei(web3.eth.getBalance(account2), 'ether')
		输出： 1
	*/

	done := s.WaitForBlockCompletation(signedTx.Hash().Hex())
	if done == 1 {

		tx, isPending, err := s.WsClient.TransactionByHash(context.Background(), signedTx.Hash())
		if err != nil {
			s.logger.Error("SendTransaction failed ", zap.Error(err))
		}
		s.logger.Info("交易完成", zap.String("交易哈希: ", tx.Hash().Hex()), zap.Bool("isPending: ", isPending))
		return nil
		
	} else {
		s.logger.Error("交易失败")
		return errors.New("交易失败")
	}

}

var ProviderSet = wire.NewSet(New, NewEthClientProviderOptions)
