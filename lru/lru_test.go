package lru

import (
	"fmt"
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("k1", String("1234")) // value要实现 Len()方法，原生的string虽然有len()方法
	if v, ok := lru.Get("k1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit k1=1234 failed")
	}
	if _, ok := lru.Get("k2"); ok {
		t.Fatalf("cache miss k2 failed")
	}
}

func TestRemoveoldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), callback)
	lru.Add("key1", String("123456"))
	fmt.Println(lru.nbytes)
	lru.Add("k2", String("k2"))
	fmt.Println(lru.nbytes)
	lru.Add("k3", String("k3"))
	fmt.Println(lru.nbytes)
	lru.Add("k4", String("k4"))
	fmt.Println(lru.nbytes)

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}
