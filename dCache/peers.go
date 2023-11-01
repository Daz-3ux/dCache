// Copyright 2023 daz-3ux(Daz) <daz-3ux@proton.me>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/Daz-3ux/dCache.

package dCache

// Picker 定义了获取分布式节点的能力
type Picker interface {
	Pick(key string) (peer Fetcher, ok bool)
}

// Fetcher 定义了从远端获取缓存的能力, 每个 Peer 都应该实现此接口
type Fetcher interface {
	Fetch(group string, key string) ([]byte, error)
}
