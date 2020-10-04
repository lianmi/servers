# see:  https://blog.csdn.net/ffzhihua/article/details/81126932?utm_medium=distribute.pc_relevant.none-task-blog-BlogCommendFromMachineLearnPai2-2.channel_param&depth_1-utm_source=distribute.pc_relevant.none-task-blog-BlogCommendFromMachineLearnPai2-2.channel_param

// 补齐64位，不够前面用0补齐
function addPreZero(num){
  var t = (num+'').length,
  s = '';
  for(var i=0; i<64-t; i++){
    s += '0';
  }
  return s+num;
}
 
web3.eth.getTransactionCount(currentAccount, web3.eth.defaultBlock.pending).then(function(nonce){
 
    // 获取交易数据
    var txData = {
        nonce: web3.utils.toHex(nonce++),
        gasLimit: web3.utils.toHex(99000),   
        gasPrice: web3.utils.toHex(10e9),
        // 注意这里是代币合约地址    
        to: contractAddress,
        from: currentAccount,
        // 调用合约转账value这里留空
        value: '0x00',         
        // data的组成，由：0x + 要调用的合约方法的function signature + 要传递的方法参数，每个参数都为64位(对transfer来说，第一个是接收人的地址去掉0x，第二个是代币数量的16进制表示，去掉前面0x，然后补齐为64位)
        data: '0x' + 'a9059cbb' + addPreZero('3b11f5CAB8362807273e1680890A802c5F1B15a8') + addPreZero(web3.utils.toHex(1000000000000000000).substr(2))
    }
 
    var tx = new Tx(txData);
 
    const privateKey = new Buffer('your account privateKey', 'hex'); 
 
    tx.sign(privateKey);
 
    var serializedTx = tx.serialize().toString('hex');
 
    web3.eth.sendSignedTransaction('0x' + serializedTx.toString('hex'), function(err, hash) {
        if (!err) {
            console.log(hash);
        } else {
            console.error(err);
        }
    });
});