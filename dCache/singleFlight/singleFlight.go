// Copyright 2023 daz-3ux(Daz) <daz-3ux@proton.me>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/Daz-3ux/dCache.

package singleFlight

import (
	"sync"
)

// call 实例代表这更在进行中，或已经结束的请求。使用 sync.WaitGroup 锁避免重入。
type call struct {
	wg  sync.WaitGroup // 一个计数信号量，可以用来记录并维护运行的goroutine
	val interface{}    // 存储任意类型的返回值
	err error
}

// Group 是 singleFlight 的主数据结构，管理不同 key 的请求 (call)
type Group struct {
	mu sync.Mutex       // 保护 m
	m  map[string]*call // 用于记录每个key的call
}

// Do 方法接收一个 key 和一个函数 fn。在函数 fn 被调用的过程中，相同 key 的其它调用都会被阻塞在 Do 方法调用处
// 确保 fn 只会被调用一次
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	// 如果 m 中存在 key，说明有其它请求正在进行，直接等待
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait() // 等待正在进行的请求
		return c.val, c.err
	}

	// 如果 m 中不存在 key，说明是第一次请求，创建一个 call，并使用 sync.WaitGroup 对 call 的 wg 计数加1
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	// 调用 fn，进行请求，并更新 call 的 val 和 err
	c.val, c.err = fn()
	c.wg.Done() // 请求结束，计数减1

	// 更新 g.m，删除 key 对应的 call
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
