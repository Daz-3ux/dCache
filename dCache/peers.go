package dCache

// PeerPicker 是必须实现的接口, 用于定位拥有特定键的对等节点
type PeerPicker interface {
    PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 是对等节点必须实现的接口, 用于从对应 group 查找缓存值
type PeerGetter interface {
    Get(group string, key string) ([]byte, error)
}
