// Copyright 2023 daz-3ux(Daz) <daz-3ux@proton.me>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/Daz-3ux/dCache.

package lru

import (
	"container/list"
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
}

// OnEvicted 记录某条记录被移除时的回调函数
type OnEvicted func(key string, value Value)

// entry 是双向链表节点的数据类型
type entry struct {
	key   string
	value Value
}

// Value 是缓存值的抽象接口，Len() 返回值所占用的内存大小
type Value interface {
	Len() int
}

func New(maxBytes int64, callback OnEvicted) *Cache {
	return &Cache{
		capacity: maxBytes,
		ll:       list.New(),
		hashmap:  make(map[string]*list.Element),
		callback: callback,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.hashmap[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
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
		ele := c.ll.PushFront(&entry{key, value})
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
