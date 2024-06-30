package singleflight

import (
	"sync"
)

type call struct { // 正在进行中或者已经结束的请求
	wg  sync.WaitGroup
	val interface{}
	err error
	dup bool
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// group.load()被并发调用，g.loader.Do()被并发调用，有一个线程第一个拿到锁，第一次new(call)
// wg.Add(1)，调用fn()后 wg.Done()，最后又删除了 map[key]（删除key的原因：占用内存了；缓存key都放在lru里面，如果这里不删除key，那么key-value还要保持更新。）
// 其他线程在第一个线程调用fn()之前，都只能 wg.Wait()等等（因为被 wg.Add(1)阻塞）
// 那如果有线程是在删除了 map[key]之后去拿到锁，g.m[key]依然不存在这样又要 new一遍 call了？

// 针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，
// 等待 fn 调用结束了，返回返回值或错误
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		c.dup = true
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	// g.mu.Lock()
	// delete(g.m, key)
	// g.mu.Unlock()
	// 1. 下面是另一种写法，上面的写法实际无法做到使fn()只被调用一次
	// 通过singleflight_test看，这是有效的。
	// go func() {
	// 	time.Sleep(time.Second)
	// 	delete(g.m, key)
	// }()
	// 2. 这种做法也可以。
	g.mu.Lock()
	if c.dup {
		delete(g.m, key) // 仅仅再确认没有重复请求时删除
	}
	g.mu.Unlock()

	return c.val, c.err
}
