# docker 搭建双节点ipfs

参考：
https://learnblockchain.cn/article/1702

# 二、开始搭建
## 1、生成swarm.key

warm.key是一个共享密钥，只有拥有相同密钥的节点才能互相通信，组成一个私钥网络。
swarm.key可以使用https://github.com/Kubuxu/go-ipfs-swarm-key-gen工具生成，
工具的安装命令是：
```
go get -u github.com/Kubuxu/go-ipfs-swarm-key-gen/ipfs-swarm-key-gen

```

```
ipfs-swarm-key-gen > /store/ipfs/swarm.key

其中

/store/ipfs/  目录是我ipfs目录。
/store/ipfs/node1 是我ipfs 节点1目录。
/store/ipfs/node2 是我ipfs 节点2目录。
```

# 2、启动节点
## 2.2.1 运行节点1和节点2

``` 
// 运行节点1
docker run -d --name ipfs_node_1 -e IPFS_SWARM_KEY_FILE=/store/ipfs/swarm.key -v /store/ipfs/node1/staging:/export -v /store/ipfs/node1/data:/data/ipfs -p 4001:4001 -p 4001:4001/udp -p 127.0.0.1:8080:8080 -p 127.0.0.1:5001:5001 ipfs/go-ipfs:latest

// 运行节点2
docker run -d --name ipfs_node_2  -e IPFS_SWARM_KEY_FILE=/store/ipfs/swarm.key -v /store/ipfs/node2/staging:/export -v /store/ipfs/node2/data:/data/ipfs -p 4002:4001 -p 4002:4001/udp -p 127.0.0.1:8081:8080 -p 127.0.0.1:5002:5001 ipfs/go-ipfs:latest


```

## 2.2.2 清除所有缺省启动节点bootstrap
```
docker exec ipfs_node_1 ipfs bootstrap rm all
docker exec ipfs_node_2 ipfs bootstrap rm all

```

## 2.2.3 查看节点id
```
docker exec ipfs_node_1 ipfs id
docker exec ipfs_node_2 ipfs id

```
输出:

```
# 节点1
{
    ...
    /ip4/172.17.0.2/tcp/4001/p2p/12D3KooWFqZtLhyN6aZKs5EsFgtpmct7HWcLE2cVZq88t252Dxcj
    ...
}
# 节点2
{
    ...
     /ip4/172.17.0.3/tcp/4001/p2p/12D3KooWGDem5ZQWmfe2m3BHKsd1VZakRTUgF38xW2ynaWJK35nH
    ...
}
```

## 2.2.4 添加节点id

在节点1中添加节点2地址
```
docker exec ipfs_node_1 ipfs bootstrap add  /ip4/172.17.0.3/tcp/4001/p2p/12D3KooWGDem5ZQWmfe2m3BHKsd1VZakRTUgF38xW2ynaWJK35nH
```

在节点2中添加节点1地址
```
docker exec ipfs_node_2 ipfs bootstrap add /ip4/172.17.0.2/tcp/4001/p2p/12D3KooWFqZtLhyN6aZKs5EsFgtpmct7HWcLE2cVZq88t252Dxcj
```

至此，我们2个节点的IPFS私有网络已搭建完成。


# 三、用一下
```
docker exec ipfs_node_1 ipfs -h
```

## 3.1 添加文件 add
```
$ docker exec ipfs_node_1 ipfs add /data/ipfs/swarm.key

 67.94 KiB / 67.94 KiB  100.00%added QmfJo1mTz3FtMdE9mTr24GbeLXdRu4Emo4VS1UpkTgXGVd swarm.key
```

## 3.2 查看文件 cat
```
$ docker exec ipfs_node_1 ipfs cat QmfJo1mTz3FtMdE9mTr24GbeLXdRu4Emo4VS1UpkTgXGVd 
$ docker exec ipfs_node_2 ipfs cat QmfJo1mTz3FtMdE9mTr24GbeLXdRu4Emo4VS1UpkTgXGVd 


```

## 3.3 下载文件 get
```
$ docker exec ipfs_node_2 ipfs get QmfJo1mTz3FtMdE9mTr24GbeLXdRu4Emo4VS1UpkTgXGVd -o /data/ipfs/test.key

```

## 3.4 查看文件列表 ls
```
$ docker exec ipfs_node_2 ipfs pin ls
```
