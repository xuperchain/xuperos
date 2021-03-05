# XuperChain XuperOS

XuperOS是XuperChain技术体系开源项目之一，是构建超级链开放网络的底层技术。基于XuperCore动态内核实现的，适用于公开网络构建的区块链发行版。您可以使用XuperOS，作为区块链基础设施，构建合规的区块链网络。

# 快速使用

## 编译环境

- Linux or Mac OS
- Go 1.13
- git、make、gcc、wget、unzip

## 部署测试网络

XuperOS提供了测试网络部署工具，初次尝试可以通过该工具便捷部署测试网络。

```
// clone项目
git clone https://github.com/xuperchain/xuperos.git
// 进入工程目录
cd xuperos
// 编译工程
make all
// 部署测试网络
make testnet
// 分别启动三个节点（确保端口未被占用）
cd ./testnet/node1
sh ./control.sh start
cd ../node2
sh ./control.sh start
cd ../node3
sh ./control.sh start
// 观察每个节点状态
./bin/xchain-cli status -H :36301
./bin/xchain-cli status -H :36302
./bin/xchain-cli status -H :36303

```

测试网络搭建完成，开启您的区块链之旅！

# 参与贡献

XuperOS目前还在建设阶段，欢迎感兴趣的同学一起参与贡献。

如果你遇到问题或需要新功能，欢迎创建issue。

如果你可以解决某个issue, 欢迎发送PR。
