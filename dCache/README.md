## 单机分布缓存
- 使用 sync.Mutex 保证并发安全
  - 封装了 LRU 的方法,使其支持并发读写

## 主体结构 Group
- Group 是 GeeCache 最核心的数据结构，负责与外部交互，控制缓存存储和获取的主流程
```go
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}
```

```
                            是
接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
                |  否                         是
                |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
                            |  否
                            |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶
```


## HTTP 服务器
- 通过 HTTP 服务器保证分布式缓存的节点间通信
  - 如果一个节点启动了 HTTP 服务器，那么这个节点就可以接受别的节点的访问请求
  - HTTPPol 是承载节点间通信的核心数据结构
```go
type HTTPPool struct {
	self     string
	basePath string
}
```

