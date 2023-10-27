package lru

import (
    "container/list"
)

/*
   LRU 缓存淘汰算法:
   key 指向 entry 的值,也指向 Cache 中的链表节点
   cache map[string]*list.Element 与 entry 中的 key string 相同
*/

// Cache 是一个 LRU 缓存，不是并发安全的
type Cache struct {
    // 容量
    maxBytes int64
    // 已使用的内存
    nBytes int64
    ll     *list.List
    cache  map[string]*list.Element
    // 可选: 记录某条记录被移除时的回调函数
    OnEvicted func(key string, value Value)
}

// entry 是双向链表节点的数据类型
type entry struct {
    key   string
    value Value
}

// Value 是缓存值的抽象接口，Len() 返回值所占用的内存大小
type Value interface {
    Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
    return &Cache{
        maxBytes:  maxBytes,
        ll:        list.New(),
        cache:     make(map[string]*list.Element),
        OnEvicted: onEvicted,
    }
}

func (c *Cache) Get(key string) (value Value, ok bool) {
    if ele, ok := c.cache[key]; ok {
        c.ll.MoveToFront(ele)
        kv := ele.Value.(*entry)
        return kv.value, true
    }
    return
}

func (c *Cache) RemoveOldest() {
    ele := c.ll.Back()
    if ele != nil {
        // 从链表中删除
        c.ll.Remove(ele)
        kv := ele.Value.(*entry)
        // 从字典中删除
        delete(c.cache, kv.key)
        c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
        if c.OnEvicted != nil {
            c.OnEvicted(kv.key, kv.value)
        }
    }
}

func (c *Cache) Add(key string, value Value) {
    if ele, ok := c.cache[key]; ok {
        c.ll.MoveToFront(ele)
        kv := ele.Value.(*entry)
        c.nBytes += int64(value.Len()) - int64(kv.value.Len())
        kv.value = value
    } else {
        ele := c.ll.PushFront(&entry{key, value})
        c.cache[key] = ele
        c.nBytes += int64(len(key)) + int64(value.Len())
    }
    for c.maxBytes != 0 && c.maxBytes < c.nBytes {
        c.RemoveOldest()
    }
}

func (c *Cache) Len() int {
    return c.ll.Len()
}
