package main

import (
    "fmt"
    "github.com/Daz-3ux/dazcache/dCache"
    "log"
    "net/http"
)

var db = map[string]string{
    "daz":     "666",
    "fakedaz": "777",
    "realdaz": "888",
}

func main() {
    getter := dCache.GetterFunc(func(key string) ([]byte, error) {
        log.Println("[SlowDB] search key", key)
        if v, ok := db[key]; ok {
            return []byte(v), nil
        }
        return nil, fmt.Errorf("%s not exist", key)
    })

    dCache.NewGroup("dCache", 2<<10, getter)

    addr := "127.0.0.1:9090"
    peers := dCache.NewHTTPPool(addr)
    log.Println("dCache is running at", addr)
    log.Fatal(http.ListenAndServe(addr, peers))
}
