// Copyright 2023 daz-3ux(Daz) <daz-3ux@proton.me>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/Daz-3ux/dCache.

package dCache

import (
    "context"
    "fmt"
    "github.com/Daz-3ux/dazCache/dCache/consistentHash"
    pb "github.com/Daz-3ux/dazCache/dCache/dCachePB"
    "github.com/Daz-3ux/dazCache/dCache/register"
    "google.golang.org/grpc"
    "log"
    "net"
    "strings"
    "sync"
)

/*
   Server
*/

// 节点间通信前缀,例如 http://example.net/_dCache/
const (
    defaultAddr     = "127.0.0.1:8024"
    defaultReplicas = 100
)

// server 实现了一个 gRPC 服务器
type server struct {
    pb.UnimplementedGroupCacheServer

    addr       string     // format: ip:port
    status     bool       // true: running / false: stop
    stopSignal chan error // 通知 etcd 停止的信号
    mu         sync.Mutex
    consHash   *consistentHash.Map
    clients    map[string]*client
}

func NewServer(addr string) (*server, error) {
    if addr == "" {
        addr = defaultAddr
    }
    if !validPeerAddr(addr) {
        return nil, fmt.Errorf("invalid peer address: %s", addr)
    }
    return &server{addr: addr}, nil
}

func validPeerAddr(addr string) bool {
    token1 := strings.Split(addr, ":")
    if len(token1) != 2 {
        return false
    }
    token2 := strings.Split(token1[0], ".")
    if token1[0] != "localhost" && len(token2) != 4 {
        return false
    }
    return true
}

func (s *server) Get(ctx context.Context, in *pb.DCacheRequest) (*pb.DCacheResponse, error) {
    group, key := in.GetGroup(), in.GetKey()
    resp := &pb.DCacheResponse{}

    log.Printf("[dCache_server %s] recv RPC request - (%s)/(%s)", s.addr, group, key)
    if key == "" {
        return resp, fmt.Errorf("key is empty")
    }
    g := GetGroup(group)
    if g == nil {
        return resp, fmt.Errorf("group %s not found", group)
    }

    view, err := g.Get(key)
    if err != nil {
        return resp, err
    }
    resp.Value = string(view.ByteSlice())
    return resp, nil
}

// Start 启动 dCache 服务
func (s *server) Start() error {
    s.mu.Lock()
    if s.status == true { // running
        s.mu.Unlock()
        return fmt.Errorf("dCache server %s already running", s.addr)
    }
    /*
       1. 设置 s.status = true, 代表服务正在运行
       2. 初始化 stop channel, 用于通知 register stop keep alive
       3. 初始化 TCP Socket 并开始监听
       4. 注册 RPC 服务到 gRPC, gRPC 开始接受 Request 并分发给 server 处理
       5. 将自己的服务注册到 etcd, client 可通过 etcd 发现服务
    */

    s.status = true
    s.stopSignal = make(chan error)

    port := strings.Split(s.addr, ":")[1]
    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        return fmt.Errorf("failed to listen: %v", err)
    }
    grpcServer := grpc.NewServer()
    pb.RegisterGroupCacheServer(grpcServer, s)

    // 注册服务到 etcd
    go func() {
        // 创建一个 etcd, 除非错误否则一直运行
        err := register.Register("dCache", s.addr, s.stopSignal)
        if err != nil {
            log.Fatalf(err.Error())
        }
        // Close channel
        close(s.stopSignal)
        // Close TCP listen
        err = lis.Close()
        if err != nil {
            log.Fatalf(err.Error())
        }
        log.Printf("[%s] Revoke service and close tcp socket ok.", s.addr)
    }()

    s.mu.Unlock()

    if err := grpcServer.Serve(lis); s.status && err != nil {
        return fmt.Errorf("failed to serve: %v", err)
    }

    return nil
}

// SetPeers 将各个远端主机 IP 加入到 Server 中
func (s *server) SetPeers(peersAddr ...string) {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.consHash = consistentHash.New(defaultReplicas, nil)
    s.consHash.Add(peersAddr...)
    s.clients = make(map[string]*client)
    for _, addr := range peersAddr {
        if !validPeerAddr(addr) {
            panic(fmt.Errorf("invalid peer address: %s", addr))
        }
        service := fmt.Sprintf("dCache/%s", addr)
        s.clients[addr] = &client{name: service}
    }
}

// Pick 根据一致性哈希选择节点
func (s *server) Pick(key string) (Fetcher, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()

    peerAddr := s.consHash.Get(key)
    // Pick itself
    if peerAddr == s.addr {
        log.Printf("[dCache_server %s] Pick itself", s.addr)
        return nil, false
    }
    log.Printf("[dCache_server %s] Pick peer %s", s.addr, peerAddr)
    return s.clients[peerAddr], true
}

// Stop 停止 dCache 服务
func (s *server) Stop() {
    s.mu.Lock()
    if s.status == false {
        s.mu.Unlock()
        return
    }

    s.stopSignal <- nil // 发送信号,停止 keepalive 信号
    s.status = false
    s.clients = nil
    s.consHash = nil
    s.mu.Unlock()
}
