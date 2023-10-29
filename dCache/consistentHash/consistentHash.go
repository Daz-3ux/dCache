// Copyright 2023 daz-3ux(Daz) <daz-3ux@proton.me>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/Daz-3ux/dCache.

package consistentHash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 映射字符数组到 uint32 (无符号 32 位整数)
type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环
	hashMap  map[int]string // 虚拟节点与真实节点的映射表，键是虚拟节点的哈希值，值是真实节点的名称
}

// New 创建一致性哈希算法 Map， 参数为虚拟节点倍数和 Hash 函数
// Hash 函数的创建使用依赖注入, 可以自定义也可以是默认的 crc32.ChecksumIEEE
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 允许传入 0 或 多个真实节点的名称, 用于添加真实节点/机器
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// 创建 m.replicas 个虚拟节点, 通过添加编号的方式区分不同虚拟节点
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	// 对所有虚拟节点的哈希值进行排序, 方便之后进行二分查找
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 通过二分查找获取第一个匹配的虚拟节点的下标 idx
	// If there is no such index, Search returns n.
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// idx 可能等于 len(m.keys), 此时应该返回 m.keys[0]，因为 m.keys 是一个环状结构
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
