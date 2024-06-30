package main

import (
	"flag"
	"fmt"
	"sync"
)

func main() {
	var port int
	var api bool
	fmt.Println(2)
	wg := sync.WaitGroup{}
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "start a api server")
	flag.Parse()
	wg.Wait()
	fmt.Println(1)
}
