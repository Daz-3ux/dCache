package dCache

import (
    "fmt"
    "github.com/Daz-3ux/dazcache/dCache/consistentHash"
    "io"
    "log"
    "net/http"
    "net/url"
    "strings"
    "sync"
)

/*
   Server
*/

// 节点间通信前缀,例如 http://example.net/_dCache/
const (
    defaultBasePath = "/_dCache/"
    defaultReplicas = 100
)

// HTTPPool 实现了一个 HTTP 对等节点的池, 用于 PeerPicker 接口
type HTTPPool struct {
    // 当前节点的基础 URL, 比如 http://example.net:8080
    self        string
    basePath    string
    mu          sync.Mutex // 保证 peers 以及 httpGetter 一致性
    peers       *consistentHash.Map
    httpGetters map[string]*httpGetter // 键 e.g. "http://10.0.0.1:8080"
}

func NewHTTPPool(self string) *HTTPPool {
    return &HTTPPool{
        self:     self,
        basePath: defaultBasePath,
    }
}

// Log 用服务器名称作为前缀打印日志
func (p *HTTPPool) Log(format string, v ...interface{}) {
    log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 处理所有 HTTP 请求
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if !strings.HasPrefix(r.URL.Path, p.basePath) {
        panic("HTTPPool serving unexpected path: " + r.URL.Path)
    }

    p.Log("%s %s", r.Method, r.URL.Path)

    // 必须是 /<basepath>/<groupname>/<key> 格式
    parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
    if len(parts) != 2 {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }

    groupName := parts[0]
    key := parts[1]

    group := GetGroup(groupName)
    if group == nil {
        p.Log("no such group: %s", groupName)
        http.Error(w, "no such group: "+groupName, http.StatusNotFound)
        return
    }

    view, err := group.Get(key)
    if err != nil {
        p.Log("error getting value for key %s: %v", key, err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/octet-stream")
    _, err = w.Write(view.ByteSlice())
    if err != nil {
        p.Log("%v", err)
    }
}

/*
   Client
*/

type httpGetter struct {
    baseURL string
}

// Get 是 PeerGetter 接口的具体实现, 用于从对应 group 查找缓存值
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
    u := fmt.Sprintf(
        "%v%v/%v",
        h.baseURL,
        url.QueryEscape(group),
        url.QueryEscape(key),
    )
    res, err := http.Get(u)
    if err != nil {
        return nil, err
    }
    defer func(Body io.ReadCloser) {
        err := Body.Close()
        if err != nil {
            log.Println(err)
        }
    }(res.Body)

    if res.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("server returned: %v", res.Status)
    }

    bytes, err := io.ReadAll(res.Body)
    if err != nil {
        return nil, fmt.Errorf("reading response body: %v", err)
    }

    return bytes, nil
}

func (p *HTTPPool) Set(peers ...string) {
    p.mu.Lock()
    defer p.mu.Unlock()

    p.peers = consistentHash.New(defaultReplicas, nil)
    p.peers.Add(peers...)
    p.httpGetters = make(map[string]*httpGetter, len(peers))
    for _, peer := range peers {
        p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
    }
}

// PickPeer 是 PeerPicker 接口的具体实现, 用于根据具体的 key 选择节点
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
    p.mu.Lock()
    defer p.mu.Unlock()

    if peer := p.peers.Get(key); peer != "" && peer != p.self {
        p.Log("Pick peer %s", peer)
        return p.httpGetters[peer], true
    }

    return nil, false
}
