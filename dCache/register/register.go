// Copyright 2023 daz-3ux(Daz) <daz-3ux@proton.me>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/Daz-3ux/dCache.

package register

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"log"
	"time"
)

// register 提供注册服务到 etcd 的功能

var (
	DefaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	}
)

// Register 注册一个服务到 etcd
func Register(service string, addr string, stop chan error) error {
	// 创建一个 etcd 客户端
	cli, err := clientv3.New(DefaultEtcdConfig)
	if err != nil {
		return err
	}
	defer func(cli *clientv3.Client) {
		err := cli.Close()
		if err != nil {
			return
		}
	}(cli)

	// 创建一个租约, 5 秒后过期
	resp, err := cli.Grant(cli.Ctx(), 5)
	if err != nil {
		return fmt.Errorf("etcd grant failed, err: %v", err)
	}
	leaseID := resp.ID

	// 注册服务
	err = etcdAdd(cli, leaseID, service, addr)
	if err != nil {
		return fmt.Errorf("etcd add failed, err: %v", err)
	}

	// 设置心跳检测, 保持租约有效
	ch, err := cli.KeepAlive(cli.Ctx(), leaseID)
	if err != nil {
		return fmt.Errorf("etcd keep alive failed, err: %v", err)
	}

	log.Printf("[%s] register success\n", addr)

	// 监听 stop 信号, 上下文关闭, 以及心跳管道消息
	for {
		select {
		// 外部信号
		case err := <-stop:
			if err != nil {
				log.Println(err)
			}
			return err
			// 外部停止, 上下文关闭
		case <-cli.Ctx().Done():
			log.Println("service closed")
			return nil
			// 心跳异常
		case _, ok := <-ch:
			if !ok {
				log.Println("keep alive channel closed")
				// 如果心跳管道关闭, 则撤销租约并返回错误
				_, err := cli.Revoke(context.Background(), leaseID)
				return err
			}
		}
	}
}

// Register 在租赁模式添加一对 kv 到 etcd
func etcdAdd(c *clientv3.Client, lid clientv3.LeaseID, service string, addr string) error {
	// 创建一个 etcd 的 endpoints.Manager, 其管理特定服务的所有终端地址
	em, err := endpoints.NewManager(c, service)
	if err != nil {
		return err
	}

	endpoint := endpoints.Endpoint{Addr: addr}

	// 将 endpoint 注册到 etcd, 并指定租约
	return em.AddEndpoint(c.Ctx(), service+"/"+addr, endpoint, clientv3.WithLease(lid))
}
