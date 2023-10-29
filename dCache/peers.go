// Copyright 2023 daz-3ux(Daz) <daz-3ux@proton.me>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/Daz-3ux/dCache.

package dCache

import pb "github.com/Daz-3ux/dazCache/dCache/dCachePB"

// PeerPicker 是必须实现的接口, 用于定位拥有特定键的对等节点
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 是对等节点必须实现的接口, 用于从对应 group 查找缓存值
type PeerGetter interface {
	Get(in *pb.DCacheRequest, out *pb.DCacheResponse) error
}
