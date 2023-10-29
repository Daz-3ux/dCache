## protobuf
- 使用 protobuf 作为数据交换格式, 进行节点间的通信
  - 轻量化
  - 与语言, 平台无关
  - 可扩展可序列化
  - 以二进制方式存储
- 生成命令
```shell
protoc --go_out=. dCachePB/dCachePB.proto
```