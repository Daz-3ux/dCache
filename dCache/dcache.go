// Copyright 2023 daz-3ux(Daz) <daz-3ux@proton.me>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/Daz-3ux/dCache.

package dCache

import (
	"fmt"
	"github.com/Daz-3ux/dazCache/dCache/singleFlight"
	"log"
	"sync"
)

/*
// 定义一个函数，用于加载数据
func loadData(key string) ([]byte, error) {
	// 在这里实现加载数据的逻辑
	return []byte("data"), nil
}

// 将 loadData 函数转换为 GetterFunc 类型
getterFunc := GetterFunc(loadData)

// 调用 GetterFunc 类型的 Fetch 方法，实际上是调用 loadData 函数
// f(key) 就是调用相应的 GetterFunc 类型的函数
data, err := getterFunc.Fetch("key")
if err != nil {
    // 处理错误
} else {
    // 使用获取到的数据
    fmt.Println(data)
}

灵活 -- 适配 -- 可插拔
*/

// Getter 是一个加载指定 key 的数据的接口
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 是一个通过函数实现 Getter 接口的类型
type GetterFunc func(key string) ([]byte, error)

// Get 实现了 Getter 接口的函数
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 是 GeeCache 最核心的数据结构，负责与外部交互，控制缓存存储和获取的主流程
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	picker    Picker
	loader    *singleFlight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 创建一个新的 Group 实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleFlight.Group{},
	}
	groups[name] = g

	return g
}

// GetGroup 返回指定名称的 Group
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()

	g := groups[name]

	return g
}

// Get 从缓存中获取指定 key 的数据
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// 从 mainCache 中查找缓存，如果存在则返回缓存值
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[dCache] hit")
		return v, nil
	}

	// 如果缓存不存在，则调用 load 方法加载
	return g.load(key)
}

func (g *Group) Update(key, value string) {
	if key == "" {
		log.Println("[dCache] key is required")
		return
	}

	// Delete the corresponding entry from the cache
	g.deleteCache(key)
}

func (g *Group) deleteCache(key string) {
	g.mainCache.delete(key)
}

func (g *Group) RegisterPeers(peers Picker) {
	if g.picker != nil {
		panic("RegisterPeerPicker called more than once")
	}

	g.picker = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	// 每个 key 只加载一次，无论是缓存还是数据库, 无论是否并发
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.picker != nil {
			if peer, ok := g.picker.Pick(key); ok {
				bytes, err := peer.Fetch(g.name, key)
				if err == nil {
					return ByteView{b: bytes}, nil
				}
				log.Printf("[dCache] Failed to get [%s] from peer, %s\n", key, err.Error())
			}
		}
		return g.getLocally(key)
	})

	if err != nil {
		return view.(ByteView), err
	}

	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	// 将数据添加到缓存中
	g.populateCache(key, value)

	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
