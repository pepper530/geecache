package main

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
)

func main() {
	byteArray := []byte{54, 51, 48}
	hexString := fmt.Sprintf("%x", byteArray)
	fmt.Println("十六进制字符串:", hexString)
	decode, _ := hex.DecodeString(hexString)
	fmt.Println(string(decode))
	decimalValue, err := strconv.ParseInt(hexString, 16, 64)
	if err != nil {
		fmt.Println("转换失败:", err)
		return
	}
	fmt.Println("十进制整数:", decimalValue)

	ex := []int{2, 4, 6, 8, 10, 14, 20, 40, 45, 1, 3, 18, 32}
	sort.Ints(ex)
	fmt.Println(ex)
	idx := sort.Search(len(ex), func(i int) bool {
		return ex[i] >= 18
	})
	fmt.Printf("i: %v,,,,val:%v", idx, ex[idx])
}
