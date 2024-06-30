package geecache

import (
	"fmt"
	"io/ioutil"
	"log"
	ch "module/consistenthash"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// http通信的服务端
type HTTPPool struct {
	self        string // 自己的地址, ip+port
	basePath    string // 节点间通讯地址的前缀。和主机上承载的其他服务区分开
	mu          sync.Mutex
	peers       *ch.Map                // 根据 key选择对应的节点（用一致性哈希）
	httpGetters map[string]*httpGetter // 节点与对应的 httpGetter一一映射。每一个远程节点对应一个 httpGetter
}

// 初始化节点的 httpPool
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// 自身节点url 与 映射的节点url
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s %s]", p.self, fmt.Sprintf(format, v...))
}

// 作为通信的Server端，实现的逻辑
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	// /basePath/groupname/key
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	bView, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// 将缓存值作为 http.Response的Body返回
	w.Header().Set("Content-Type", "application/octet-stream") // 二进制流
	w.Write(bView.ByteSlice())                                 // content是字节数组

}

// -------------------- 下面是 http Client的实现
// http通信的客户端
type httpGetter struct {
	baseURL string
}

// 通信的客户端类 httpGetter，实现 PeerGetter接口
// 从group查找key的只读缓存: 调用 http.Get(url)
func (h *httpGetter) Get(groupName string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(groupName), // groupName
		url.QueryEscape(key),
	)
	log.Printf("httpGetter | Get from url: %v \n", u)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	log.Println("httpGetter | http.Get(u) success, return bytes")
	return bytes, nil
}

// 这种定义方式，用来检测 httpGetter是否实现了 PeerGetter接口，
// 如果没有实现，那么源码编译时则会报错
var _ PeerGetter = (*httpGetter)(nil)

// 实例化一致性哈希算法，添加传入真实节点，为每一个节点创建一个http客户端 httpGetter
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = ch.New(defaultReplicas, nil) // p.peers是一致性哈希数据结构 Map（初始化）
	p.peers.Add(peers...)                  // 传入的peers就是真实节点url，Add之后创建了带虚拟节点的哈希环
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers { // peer: http://localhost:8001
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
		// log.Printf("httppool | Set peers: %v, %v\n", p.httpGetters, p.peers)
		log.Printf("httppool | Set peers: %v | baseURL: %v\n", p.httpGetters, peer+p.basePath)
	}
}

// 实现 PickPeer接口，根据具体的 key选择节点，返回节点对应的 http客户端
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// p.peers是哈希环
	// 获取真实节点名称 peer，是远程节点，不能是本机 p.self
	peer := p.peers.Get(key)
	log.Printf("httppool | PickPeer: %v, %v", peer, p.self)
	if peer != "" && peer != p.self {
		p.Log("Pick peer: %s", peer)
		return p.httpGetters[peer], true // 返回真实节点的httpGetter客户端
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)
