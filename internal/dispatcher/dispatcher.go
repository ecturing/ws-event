package dispatcher

import (
	"hash/fnv"
	"runtime"
	"ws-event/internal/poller"
	eventif "ws-event/pkg/eventIF"

	"github.com/gorilla/websocket"
)

// poller dispatcher,事件循环调度器，负责把conn分发到各个不同的eventloop中

var (
		cpus       = runtime.NumCPU()
)

type Dispatcher interface {
	Dispatch(conn *websocket.Conn) *poller.Multiplexer
	run()
}

type HashDispatcher struct {
	PollerList []*poller.Multiplexer
}

func Default() HashDispatcher {
	return HashDispatcher{
		PollerList: make([]*poller.Multiplexer, cpus),
	}
}

func UseDispatcher(d Dispatcher) Dispatcher{
	d.run()
	return d
}

func (h HashDispatcher) run() {
	for i := range h.PollerList {
		h.PollerList[i] = poller.Poller_create()
		go h.PollerList[i].Poll()
	}
}


// 根据hash数值分发
func (d HashDispatcher) Dispatch(conn *websocket.Conn) *poller.Multiplexer{
		tuple := conn.RemoteAddr().String() + conn.LocalAddr().String()
		h := fnv.New32a()
		h.Write([]byte(tuple))
		result := h.Sum32()
		return d.PollerList[int(result) % cpus]

}


func (d HashDispatcher) RegisterConn(conn *websocket.Conn, event eventif.WsEvent) {
	upgradeConn := poller.WsEnchent_create(conn)
	poller:=d.Dispatch(conn)
	upgradeConn.PollerSelect(poller)
	upgradeConn.Poller.Poller_add(upgradeConn)
	upgradeConn.RegisterHandler(event)
}
