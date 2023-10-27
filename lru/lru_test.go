package lru

import (
    "testing"
)

type String string

func (d String) Len() int {
    return len(d)
}

func TestCache_Get(t *testing.T) {
    // 0 代表不限制内存大小
    lru := New(int64(0), nil)
    lru.Add("key1", String("1234"))
    if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
        t.Fatalf("cache hit key1=1234 failed")
    }
    if _, ok := lru.Get("key2"); ok {
        t.Fatalf("cache miss key2 failed")
    }
}

func TestCache_RemoveOldest(t *testing.T) {
    k1, k2, k3 := "key1", "key2", "k3"
    v1, v2, v3 := "value1", "value2", "3"
    capacity := len(k1 + k2 + v1 + v2)
    lru := New(int64(capacity), nil)
    lru.Add(k1, String(v1))
    lru.Add(k2, String(v2))
    lru.Add(k3, String(v3))
    if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
        t.Fatalf("Removeoldest key1 failed")
    }
}

func TestCache_TestOnEvicted(t *testing.T) {

}
