# 区块链
以ethereum为区块链
发行ERC20连米币

# 参考单节点
/Users/mac/developments/lianmi/blockchain/ethereum-simplenode-erc20/simplenode

```
Your new key was generated

Public address of the key:   0x9c68b2493DFD89F0A52Cbd827Cc9fbf683d56FE7
Path of the secret key file: node0/keystore/UTC--2020-09-20T11-45-09.387467825Z--9c68b2493dfd89f0a52cbd827cc9fbf683d56fe7

- You can share your public address with anyone. Others need it to interact with you.
- You must NEVER share the secret key with anyone! The key controls access to your funds!
- You must BACKUP your key file! Without the key, it's impossible to access account funds!
- You must REMEMBER your password! Without the password, it's impossible to decrypt the key!
```

# 运行 
```
$ docker-compose up -d
```

# 进入console
```
$ geth attach http://localhost:8545
```

# 初始账号及地址
```
> eth.accounts
["0xe0380828902269bfbce6b056ae3bfce8d52fd6a8", "0xf490774d9b87f4d379c2a789e5755156c1c370bc", "0xa7cc1ae7199cce8aa1354059953f6559cf57869f", "0x9c68b2493dfd89f0a52cbd827cc9fbf683d56fe7"]
```