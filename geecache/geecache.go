package geecache

import (
	"fmt"
	"log"
	"module/singleflight"
	"sync"
)

// 缓存中查不到时，就要去数据源（文件或者数据库）查找。
// 但是，怎么从数据源查找获取该数据是用户的事情。
// 所以，在这里只设计一个回调函数，缓存中找不到时就调用该回调函数去数据源查找

// Getter接口，包含回调函数Get(key string) ([]byte, error)
type Getter interface {
	Get(key string) ([]byte, error)
}

// 定义一个GetterFunc的函数类型
type GetterFunc func(key string) ([]byte, error)

// 该函数类型实现某一个接口，称之为接口型函数，
// 方便使用者在调用时既能够传入函数作为参数，也能够传入实现了该接口的结构体作为参数。
// GetterFunc实现了Getter接口
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key) // 调用 f 自己本身
}

// Group是GEE_CACHE里面最重要的结构。
// 一个Group可以被认为是一个缓存的命名空间，
// 可以创建很多Group，他们有自己唯一的名字name
type Group struct {
	name      string     // 缓存命名空间
	getter    Getter     // 缓存未命中时的回调
	mainCache cache      // 并发缓存
	peers     PeerPicker //节点选择器
	loader    *singleflight.Group
}

var (
	mu     sync.RWMutex // 这里使用读写锁，而没有用 sync.Mutex，看看两者区别？
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Group的Get()方法，返回的是只读结构的缓存值
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok { // 如果本地的mainCache中有，直接返回
		log.Printf("geecache | get from local cache: %v \n", v)
		return v, nil
	}
	return g.load(key) // 如果缓存未被命中，要从数据源获取；或者从分布式环境中的其他节点获取
}

// 先通过 PickPeer选择节点，
// singleflight实现的 Do()方法，使得并发调用 Do()时，匿名函数只被调用一次
func (g *Group) load(key string) (val ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (interface{}, error) { // Do的第二个参数是个匿名函数，能返回interface和error就行
		log.Printf("geecache | group.load() : g.name: %v, g.peers: %v", g.name, g.peers)
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if val, err = g.getFromPeer(peer, key); err == nil {
					log.Println("[GetCache] Success to get byteview from peer: ", val)
					return val, nil
				}
				log.Println("[GetCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// 实现 PeerGetter接口的 httpGetter从访问远程节点，获取缓存值
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key) // httpGetter.Get()
	if err != nil {
		return ByteView{}, err
	}
	log.Printf("geecache | getFromPeer: %v, %v\n", bytes, g.name)
	return ByteView{b: bytes}, nil
}

// 主要是调用用户的回调函数（从数据源获取数据）
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key) // 用户的回调函数
	if err != nil {                 // 回调去数据源查也没有查到
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	log.Printf("geecache | getLocally: get from getter")
	return value, nil
}

// 将源数据添加到本地mainCache缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
	log.Println("geecache | populateCache: add value to local cache")
}

// 将实现了PeerPicker接口的 HTTPPool注入到 Group中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
	log.Printf("geecache | RegisterPeers: %v \n", g.peers)
}
