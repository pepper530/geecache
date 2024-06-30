package geecache

// 只读的数据结构，存储真实的缓存值
type ByteView struct {
	b []byte
}

// 被缓存对象必须实现Value()接口，也就是Len()方法
func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (v ByteView) String() string {
	return string(v.b)
}
