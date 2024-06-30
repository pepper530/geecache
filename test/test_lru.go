package main

import (
	"fmt"
	"module/lru"
)

type String string

func (s String) Len() int {
	return len(s)
}

func main() {
	lru := lru.New(int64(0), nil)
	lru.Add("k1", String("1234")) // value要实现 Len()方法，原生的string虽然有len()方法
	v, ok := lru.Get("k1")
	fmt.Println(v)
	fmt.Println(ok)
	v2, ok := lru.Get("k2")
	fmt.Println(v2, ok)
}
