// Copyright 2023 daz-3ux(Daz) <daz-3ux@proton.me>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/Daz-3ux/dCache.

package lru

import (
	"container/list"
	"encoding/json"
	"os"
	"time"
)

/*
   LRU 缓存淘汰算法:
   key 指向 entry 的值,也指向 Cache 中的链表节点
   hashmap map[string]*list.Element 与 entry 中的 key string 相同
*/

// Cache 是一个 LRU 缓存，不是并发安全的
type Cache struct {
	// 容量
	capacity int64
	// 已使用的内存
	nBytes   int64
	ll       *list.List
	hashmap  map[string]*list.Element
	callback OnEvicted
	K        int           // 最近 K 次访问
	TTL      time.Duration // 生存时间
}

// OnEvicted 记录某条记录被移除时的回调函数
type OnEvicted func(key string, value Value)

// entry 是双向链表节点的数据类型
type entry struct {
	key         string
	value       Value
	accessTimes []time.Time
	expireAt    time.Time
}

// Value 是缓存值的抽象接口，Len() 返回值所占用的内存大小
type Value interface {
	Len() int
}

type Config struct {
	K   int           `json:"k"`
	TTL time.Duration `json:"TTL"`
}

func readConfig() (Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return Config{}, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func New(maxBytes int64, callback OnEvicted) *Cache {
	config, err := readConfig()
	if err != nil {
		return nil
	}
	ttl, err := time.ParseDuration(config.TTL.String())
	if err != nil {
		return nil
	}
	return &Cache{
		capacity: maxBytes,
		ll:       list.New(),
		hashmap:  make(map[string]*list.Element),
		callback: callback,
		K:        config.K,
		TTL:      ttl,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.hashmap[key]; ok {
		kv := ele.Value.(*entry)
		if kv.expireAt.Before(time.Now()) {
			c.Delete(key)
			return nil, false
		}
		kv.accessTimes = append(kv.accessTimes, time.Now())
		if len(kv.accessTimes) > c.K {
			kv.accessTimes = kv.accessTimes[1:]
			c.ll.MoveToFront(ele)
		}
		return kv.value, true
	}
	return
}

func (c *Cache) Delete(key string) bool {
	if ele, ok := c.hashmap[key]; ok {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.hashmap, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.callback != nil {
			c.callback(kv.key, kv.value)
		}
		return true
	}
	return false
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		// 从链表中删除
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		// 从字典中删除
		delete(c.hashmap, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.callback != nil {
			c.callback(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.hashmap[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		expireAt := time.Now().Add(c.TTL)
		ele := c.ll.PushFront(&entry{
			key,
			value,
			[]time.Time{time.Now()},
			expireAt})
		c.hashmap[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.capacity != 0 && c.capacity < c.nBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
