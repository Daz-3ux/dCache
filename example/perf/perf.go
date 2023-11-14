// Copyright 2023 daz-3ux(Daz) <daz-3ux@proton.me>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/Daz-3ux/dCache.

package main

import (
	"fmt"
	"github.com/Daz-3ux/dazCache/dCache"
	"log"
)

func main() {
	// 模拟MySQL数据库
	var mysql = map[string]string{
		"daz":     "666",
		"realdaz": "777",
		"fakedaz": "888",
	}
	// 新建cache实例
	group := dCache.NewGroup("scores", 2<<10, dCache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[MySQL] search key", key)
			if v, ok := mysql[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	// 启动一个服务实例
	var addr string = "localhost:8088"
	svr, err := dCache.NewServer(addr)
	if err != nil {
		log.Fatal(err)
	}
	// 设置同伴节点 IP (包括自己)
	svr.SetPeers(addr)
	// 将服务与 cache 绑定 因为 cache 和 server 是解耦合的
	group.RegisterPeers(svr)
	log.Println("dCache is running at", addr)

	// 启动服务(注册服务至 etcd / 计算一致性哈希)
	go func() {
		// Start将不会return 除非服务stop或者抛出error
		err = svr.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	_, err = group.Get("realdaz")
	if err != nil {
		return
	}

	for i := 0; i < 100000; i++ {
		_, err := group.Get("realdaz")
		log.Printf("get realdaz...")
		if err != nil {
			log.Println(err.Error())
			return
		}
	}
}
