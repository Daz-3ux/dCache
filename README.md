# dCache
dCache 是一个基于 Go 开发的`分布式缓存系统`  
使用 `Go` + `gRPC` +  + ``


## Features
- 使用 [LRU](./dCache/lru/README.md) 进行缓存淘汰
- 使用 [一致性哈希](./dCache/consistentHash/README.md) 进行节点选择
- 实现 [分布式节点](./dCache/README.md) 选择
- 使用 [单机并发控制](./dCache/singleFlight/README.md) 防止缓存击穿
- 使用 [gRPC](./dCache/dCachePB/README.md) 作为通信协议
- 使用 [Protobuf](./dCache/dCachePB/README.md) 作为序列化方

## Installation
- 构建
```shell
make run
```
- 测试
```shell
make test
```
- 运行
```shell
make run
```

## 架构
![架构](./assert/arch.png)

- 流程
```
                            是
接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
                |  否                         是
                |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
                            |  否
                            |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶
```

## License
[MIT](https://choosealicense.com/licenses/mit/)

