## etcd
- etcd 是一个高可用强一致性的分布式 key-value 存储系统
  - 用于共享配置和服务发现
  - 本质上来说,服务发现就是想要了解集群中是否有进程在监听 UDP 或 TCP 端口,并且通过名字就可查找其并建立连接

## 引入 etcd
- 注册方法
  - New 一个 etcd 的 Client
  - 用 Grant 方法创建一个租约
    - 默认时间为 5 秒 
  - 将服务节点连同租约一起注册到 etcd
  - 通过 KeepAlive 方法, 得到一个 chan
  - 用 for ... select 进行 channel 监听
    - 监听 外部中断
    - 监听 上下文关闭
    - 监听 心跳异常

- 发现方法
  - 通过 etcd 客户端实现的 Dial 方法监听事件

## Raft 算法
- Raft 中的 Term
  - 任期
  - 一次竞选出 leader 到下一轮竞选的时间
  - 如果 follower 接收不到 leader 的心跳信息,就会结束当前 term ,变为 candidate 继而发起竞选,帮
    助 leader 故障时集群的恢复
  - 如果集群不出现故障,那么一个 term 将无限延续下去
- Raft 状态机的切换
  - Raft 有三种状态: leader, follower, candidate
  - Node 默认进入 Follower 状态, 等待 Leader 发送心跳信息
    - 若等待超时就由 Follower 转换到 Candidate 状态进入下一轮 Term 进行 leader 选举
    - 收到集群中多数节点的投票时,该节点就转变为 Leader 状态
