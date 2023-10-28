package dCache

// ByteView 持有一个只读的字节数组, 表示缓存值
type ByteView struct {
    b []byte
}

func (v ByteView) Len() int {
    return len(v.b)
}

// ByteSlice 返回一个拷贝, 防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
    return cloneBytes(v.b)
}

// String 返回缓存值的字符串类型, 在必要时进行拷贝
func (v ByteView) String() string {
    return string(v.b)
}

func cloneBytes(b []byte) []byte {
    c := make([]byte, len(b))
    copy(c, b)
    return c
}
