package lru

import (
	"container/list"
)

// 包含字典和双向链表
type Cache struct {
	maxBytes int64 // 允许使用的最大内存
	nbytes   int64 // 当前已经使用的内存，这个内存大小是包含ll节点的key的大小和节点的value的大小的
	cache    map[string]*list.Element
	// value是链表中某个节点的指针. 另外，list的Element的Value字段是any类型，实际是interface{}空接口类型
	ll        *list.List                    // 双向链表存的才是真正的值，每个节点entry的value存值，entry的key就是map的key
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数
}

// 双向链表的节点
type entry struct {
	key   string
	value Value // value可以是任意类型，只要他实现了 Value接口（ 这里就是为了通用性）
}

// 计算value占用了多少内存
type Value interface {
	Len() int
}

// 实例化Cache， 允许的最大内存和删除时的回调是自己传入。
func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		nbytes:    0,
		cache:     make(map[string]*list.Element),
		ll:        list.New(),
		OnEvicted: onEvicted,
	}
}

// 查找功能： 从map中找到链表里的目标节点，将该节点移动到队尾(高频次访问)
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok { // ele是*list.Element类型，是某个节点的指针
		c.ll.MoveToFront(ele) // 双向链表的头和尾是相对的，这里作者定义Front为尾了，为了后面的统一，这里不按自己的理解改了
		// ele.Value 是interface{} 类型，如果不进行下面的类型断言，是不能访问到 *entry的value的
		kv := ele.Value.(*entry) // ele就是map的value:list.Element。 *entry是类型断言，告诉编译器 ele.Value 实际上应该被视为 *entry 类型
		return kv.value, true    // 这里不能转成 entry，必须是 *entry, 因为ll就是*list.Element，指针类型
	}
	return
}

// 删除最近最少被访问的节点，也就是队首的节点。
// 同时要删除map里面的key，并更新Cache结构目前的内存大小
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() // Back()返回的是 last element(对应作者定义的“头”),  Front()返回的是 first element
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key) // 还要删除 map里面的 key
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// 如果key存在，就更新value，更新占用内存大小
// 如果不存在，就添加到链表队尾（高频区），建立map[key]和节点的映射关系，更新占用内存大小
// 判断一下新内存是否超过了maxBytes，超过了就要做删低频访问节点的操作
// 这里有一个小疑惑？？传入的 key应该是什么？我们的缓存系统应该是只care链表里面存的值，这才是目标存储值，
// 或者key就设计者随意定义赋值了，只要他能和目标存储的entry能建立映射就 ok了？
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len()) // 更新占用内存大小
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}

}

func (c *Cache) Len() int {
	return c.ll.Len()
}
