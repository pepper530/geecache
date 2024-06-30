package main

import (
	"flag"
	"fmt"
	"log"
	"module/geecache"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// 这里实例化 Group时给了 groupname、回调Getter(就是本地查找时调用的Get方法)
func createGroup() *geecache.Group {
	return geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	peerserver := geecache.NewHTTPPool(addr) // peerserver就是 httppool实例
	peerserver.Set(addrs...)                 // 这里是传入所有节点url
	gee.RegisterPeers(peerserver)            // peerserver也是PeerPicker，因为httppool实现了 PickPeeer方法
	log.Println("geecache is running at: ", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peerserver))
}

func startAPIServer(apiAddr string, gee *geecache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil)) // 用nil和用指定的实现了ServeHTTP方法的http.Handler的区别？

}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v) // 就是上面的三个url
	}

	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee) // 这里为gee加入了peers(PeerPicker)
	}
	startCacheServer(addrMap[port], []string(addrs), gee)
}
