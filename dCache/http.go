package dCache

import (
    "fmt"
    "log"
    "net/http"
    "strings"
)

// 节点间通信前缀,例如 http://example.net/_dCache/
const defaultBasePath = "/_dCache/"

// HTTPPool 实现了一个 HTTP 对等节点的池, 用于 PeerPicker 接口
type HTTPPool struct {
    // 当前节点的基础 URL, 比如 http://example.net:8080
    self     string
    basePath string
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
