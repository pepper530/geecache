package main

import (
	"fmt"
	"log"
	"module/geecache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {

	loadCounts := make(map[string]int, len(db)) // 统计某个key调用回调函数的次数
	gee := geecache.NewGroup(
		"scores",
		2<<10,
		geecache.GetterFunc(
			func(key string) ([]byte, error) {
				log.Println("[SlowDB] search key", key)
				if v, ok := db[key]; ok {
					if _, ok := loadCounts[key]; !ok {
						loadCounts[key] = 0
					}
					loadCounts[key] += 1
					return []byte(v), nil
				}
				return nil, fmt.Errorf("%s not exists", key)
			}),
	)

	for k, v := range db {
		if view, err := gee.Get(k); err != nil || view.String() != v {
			log.Fatal("failed to get value of Tom")
		}
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 { // loadCounts的每个key应该为1，否则就是多次调用了回调函数，
			log.Fatal("cache %s miss", k)
		}
	}
	fmt.Println(loadCounts)
	if view, err := gee.Get("Tom"); err != nil {
		fmt.Println("Tom key not exists")
	} else {
		fmt.Println("find Tom key, %v", view)
	}
	fmt.Println(loadCounts)
	if view, err := gee.Get("unkown"); err == nil {
		log.Fatal("the value of unkown should be empty, but %s got", view)
	} else {
		fmt.Println("print err of unkown key: %v", err)
	}
}
