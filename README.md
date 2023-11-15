# dCache
dCache 是一个基于 Go 开发的`分布式缓存系统`, 是一个开箱即用的 server 组件   
使用 `Go` + `gRPC` + `etcd` + `singleFlight`


## Features
- 使用 [LRU-K](./dCache/lru/README.md) 进行缓存淘汰
- 使用 [一致性哈希](./dCache/consistentHash/README.md) 进行节点选择
- 使用 [singleFlight](./dCache/singleFlight/README.md) 防止缓存雪崩与缓存击穿
- 使用 [gRPC](./dCache/dCachePB/README.md) 实现节点间通信
  - 使用 [Protobuf](./dCache/dCachePB/README.md) 作为序列化方式
- 使用 [etcd](./dCache/register/README.md) 进行服务发现

## Installation
- 构建
```shell
make run
```
- 测试
  - 测试前确保本地已启动 etcd 服务
  - `docker run --name etcd -p 2379:2379 -e ALLOW_NONE_AUTHENTICATION=yes -d binami/etcd:latest`
```shell
make test
```

- 使用:  
借助 Go Modules 然后:
```go
import "github.com/peanutzhen/peanutcache"
```

## 接口
- dCache 通过封装 `Group` 对外提供服务
  - 只提供一个接口
    - Get: 从缓存中获取值
  - Group 也是命名空间, 不同的命名空间之间是相互隔离的

## 性能分析
- 测试代码见[example](./example)
- 测试环境:
  - CPU: Intel(R) Core(TM) i5-9300H CPU @ 2.40GHz
  - GPU: NVIDIA GeForce GTX 1660Ti
  - 内存: 8G+8G
  - 内核版本: 6.5.9-arch2-1
- 在缓存均命中的情况下, `go test -bench=".*"` 的结果如下:
```shell
  719850	      1545 ns/op
PASS
ok  	github.com/Daz-3ux/dazCache/example/benchMark	1.135s
```
- 缓存均命中情况下: perf 测试结果如下
  - 使用 `make perf` 进行测试
```shell
 Performance counter stats for '/home/realdaz/z-project/dCache/example/perf/perfTest':

            283.35 msec task-clock:u                     #    1.026 CPUs utilized
                 0      context-switches:u               #    0.000 /sec
                 0      cpu-migrations:u                 #    0.000 /sec
             1,518      page-faults:u                    #    5.357 K/sec
       348,824,662      cycles:u                         #    1.231 GHz
       538,245,341      instructions:u                   #    1.54  insn per cycle
       106,920,921      branches:u                       #  377.352 M/sec
         1,020,084      branch-misses:u                  #    0.95% of all branches

       0.276101514 seconds time elapsed

       0.069122000 seconds user
       0.213862000 seconds sys
```

## 流程
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

