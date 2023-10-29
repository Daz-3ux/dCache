package main

import (
    "flag"
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

func createGroup() *dCache.Group {
    getter := dCache.GetterFunc(func(key string) ([]byte, error) {
        log.Println("[SlowDB] search key", key)
        if v, ok := db[key]; ok {
            return []byte(v), nil
        }
        return nil, fmt.Errorf("%s not exist", key)
    })

    return dCache.NewGroup("dCacheTest", 2<<10, getter)
}

func startCacheServer(addr string, addrs []string, gee *dCache.Group) {
    peers := dCache.NewHTTPPool(addr)
    peers.Set(addrs...)
    gee.RegisterPeers(peers)
    log.Println("dCache is running at", addr)
    log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, dc *dCache.Group) {
    http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        key := r.URL.Query().Get("key")
        view, err := dc.Get(key)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/octet-stream")
        _, err = w.Write(view.ByteSlice())
        if err != nil {
            return
        }
    }))
    log.Println("frontend server is running at", apiAddr)
    log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
    var port int
    var api bool
    flag.IntVar(&port, "port", 8001, "dCache server port")
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
        addrs = append(addrs, v)
    }

    daz := createGroup()
    if api {
        go startAPIServer(apiAddr, daz)
    }
    startCacheServer(addrMap[port], addrs, daz)
}
