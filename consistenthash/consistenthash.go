package consistenthash

import (
	"hash/crc32"
	"log"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环, 有序的, 存的是所有节点的 hashInt
	hashMap  map[int]string // 虚拟节点哈希值与真实节点的映射表，k是虚拟节点哈希值，v是真实节点的名称
}

// 可以自定义虚拟节点倍数和 Hash函数, 这里默认使用 crc32.ChecksumIEEE
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

// 在哈希环上填充真实节点和虚拟节点
func (m *Map) Add(keys ...string) {
	log.Println("consistenthash: keys: ", keys)
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ { // 每个真实节点 key，创建 m.replicas个虚拟节点
			// 这里是把 string -> []byte -> uint32 -> int
			hashInt := int(m.hash([]byte(strconv.Itoa(i) + key))) // 虚拟节点名称：strconv.Itoa(i)+key  注意这是string拼接
			m.keys = append(m.keys, hashInt)
			m.hashMap[hashInt] = key // 关联虚拟节点哈希值和真实节点名称
		}
	}
	sort.Ints(m.keys) // int数组排序
	// m.hashMap[hashInt] = key (key就是真实url)
}

// 得到真实的节点名称
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hashInt := int(m.hash([]byte(key))) // key的 hash值，映射在哈希环上
	// 通过顺时针查找最近的一个虚拟节点的下标。
	// 是第一个比给定 hash值大的节点
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hashInt
	})
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
