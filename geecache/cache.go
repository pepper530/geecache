package geecache

import (
	"module/lru"
	"sync"
)

type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64 // 最大内存
}

// 封装Get()和Add()方法
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok { // v是entry.value
		return v.(ByteView), ok
	}
	return
}

// add时，加入的是ByteView类型
// 这里用到了延迟初始化（lazy initializtion)， 就是对象的创建是在第一次使用该对象时
// 延迟初始化是为了提高性能，减少程序内存要求
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}
