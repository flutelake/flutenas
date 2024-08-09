package terminal

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"flutelake/fluteNAS/pkg/module/flog"

	"github.com/gorilla/websocket"
)

type WebTerminal struct {
	// 超时时间，单位秒
	timeout int64
	// websocket 连接池
	conns map[string]*TerminalConn

	lock sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWebTerminal(timeout int) *WebTerminal {
	// 检查Record目录是否创建
	_, err := os.Stat(ReocrdFilePathPrefix)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(ReocrdFilePathPrefix, 0o644)
		}
	}
	return &WebTerminal{
		timeout: int64(timeout),
		conns:   make(map[string]*TerminalConn, 0),
		lock:    sync.Mutex{},
	}
}

// 启动websocket服务
func (t *WebTerminal) Start(stopCh <-chan struct{}) {
	// 定时清理未活跃的websocket连接
	// go wait.Until(t.cleaner, time.Minute*5, stopCh)

	// 清理缓存文件
	go recCleanerGo()
}

func (t *WebTerminal) updateConn(token string, conn *TerminalConn) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.conns[token] = conn
}

func (t *WebTerminal) deleteConn(token string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	c, ok := t.conns[token]
	if ok {
		c.Close()
		delete(t.conns, token)
	}
}

type CreateTerminalParam struct {
	Hostname           string
	BrowserFinderPrint string
	SourceIP           string
	User               string
	TerminalName       string
	Host               Host
}

func (t *WebTerminal) CreateTerminal(param CreateTerminalParam) (string, error) {
	uniqueID := fmt.Sprintf("%s_%s_%s_%s", param.Hostname, param.BrowserFinderPrint, param.SourceIP, param.User)
	// 随机数
	r := rand.Int31n(9999)
	src := uniqueID + "_" + param.TerminalName + fmt.Sprintf("%d", r)
	h := md5.New()
	h.Write([]byte(src))
	hstr := h.Sum(nil)
	token := hex.EncodeToString(hstr)

	// 判断名称是否重复
	for key, c := range t.conns {
		if c.name == param.TerminalName && c.uniqueID == uniqueID {
			// 通知旧的窗口 被挤下线了
			if c.conn != nil {
				c.conn.WriteMessage(websocket.BinaryMessage, []byte("A new terminal window has been opened, and this connection has been closed."))
				c.conn.Close()
				c.conn = nil
			}
			t.deleteConn(key)
			t.updateConn(token, c)
			// return "", fmt.Errorf("terminal name repeated")
			return token, nil
		}
	}

	c := newTerminalConn(nil, &param.Host, uniqueID, param.TerminalName)
	t.updateConn(token, c)
	return token, nil
}

func (t *WebTerminal) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	conn, ok := t.conns[token]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if conn.conn != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		flog.Errorf("Upgrade error: %v", err)
		return
	}
	defer t.deleteConn(token)
	conn.conn = wsConn
	t.updateConn(token, conn)

	if err := conn.Run(); err != nil {
		conn.conn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
		flog.Errorf("handler terminal websocket error: %v", err)
	}
}

// 清理过期的连接
func (t *WebTerminal) cleaner() {
	t.lock.Lock()
	defer t.lock.Unlock()

	currentTime := time.Now().Unix()
	for i, conn := range t.conns {
		if conn.lastHeatbeat+t.timeout < currentTime {
			if conn.conn != nil {
				conn.conn.Close()
			}
			delete(t.conns, i)
		}
	}
}
