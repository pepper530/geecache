package main

import (
	"fmt"
	"log"
	"module/consistenthash"
	"strconv"
)

func main() {
	// 自定义 Hash函数，直接把字符串转成对应的数字
	hash := consistenthash.New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})
	// 有 2/4/6 三个真实节点，
	// 对应的虚拟节点的哈希值是 02/12/22、04/14/24、06/16/26
	hash.Add("6", "4", "2")
	// fmt.Println(hash)
	// key对应的真实节点
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			log.Printf("Asking for [key]: %s, but yielded [node]: %s", k, v)
		}
	}
	fmt.Println(hash)
	// 这里可以思考一下，是不是增加节点只能在真实节点后面顺延，
	// 还是真实节点的中间也可以添加（比如“3”），不会导致大量映射关系失效
	// 应该后者是可以的吧。
	hash.Add("8")
	testCases["27"] = "8"
	for k, v := range testCases {
		if hash.Get(k) != v {
			log.Printf("Asking for [key]: %s, but yielded [node]: %s", k, v)
		}
	}
	fmt.Println(hash)
}
