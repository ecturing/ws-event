package poller

import (
	// "fmt"
	"bytes"
	"fmt"
	"log/slog"
	"net"
	// "sync"
	eventif "ws-event/pkg/eventIF"

	"github.com/gorilla/websocket"
	"golang.org/x/sys/unix"
)


type Multiplexer struct {
	pollerMap map[int]*WsEnchentConn
	// pollerMap *sync.Map
	epoller   int
}

type WsEnchentConn struct {
	Conn         *websocket.Conn
	EventHandler eventif.WsEvent
	Poller       *Multiplexer
	fd           int
}

type EventMsg struct {
	Data *bytes.Buffer
}

func (p *Multiplexer) Poll() {
	defer func() {
		if err := recover(); err != nil {
			v, ok := err.(error)
			if !ok {
				slog.Any("err", err)
			}
			slog.Error("deal error", slog.String("err", v.Error()))
		}
	}()

	var waitList = make([]unix.EpollEvent, 10)
	for {
		n, waiterr := unix.EpollWait(p.epoller, waitList, -1)
		if waiterr != nil {
			slog.Error("wait error", slog.String("err", waiterr.Error()))
			continue
		}
		for i := 0; i < n; i++ {
			event := waitList[i]
			slog.Info("event:", slog.Any("eventInfor:", fmt.Sprintf("eventfd:%d,event:%d", event.Fd, event.Events)))
			var wse *WsEnchentConn
			if a, ok := p.pollerMap[int(event.Fd)]; ok {
				wse = a
			}
			
			// if a, ok := p.pollerMap.Load(int(event.Fd)); ok {
			// 	wse = a.(*WsEnchentConn)
			// }
			
			if event.Events&unix.EPOLLIN != 0 {
				mstype, raw, err := wse.Conn.ReadMessage()
				data := bytes.NewBuffer(raw)
				if _, ok := err.(*websocket.CloseError); ok {
					wse.Conn.Close()
					p.Poller_del(wse)
					continue
				} else if err != nil {
					slog.Error("read error", slog.String("err", err.Error()))
					wse.EventHandler.OnError(data, err)
					wse.EventHandler.OnClose(data)
					wse.Conn.Close()
					p.Poller_del(wse)
				}
				switch mstype {
				case websocket.PongMessage:
					wse.Conn.PongHandler()
				case websocket.PingMessage:
					wse.Conn.PingHandler()
				case websocket.TextMessage:
					wse.EventHandler.OnMessage(data)
				case websocket.BinaryMessage:
					wse.EventHandler.OnMessage(data)
				}
			} else if event.Events&unix.EPOLLHUP != 0 {
				wse.Conn.Close()
				p.Poller_del(wse)
			} else if event.Events&unix.EPOLLERR != 0 {
				wse.Conn.Close()
				p.Poller_del(wse)
			}
		}
	}
}

func (p *Multiplexer) Poller_add(c *WsEnchentConn) {
	err:=unix.EpollCtl(p.epoller, unix.EPOLL_CTL_ADD, c.fd, &unix.EpollEvent{
		Events: uint32(unix.EPOLLIN | unix.EPOLLHUP | unix.EPOLLERR | unix.EPOLLET),
		Fd:     int32(c.fd),
	})
	if err != nil {
		slog.Error("add conn error", slog.String("err", err.Error()))
		return
	}	
	
	p.pollerMap[c.fd]=c
	// p.pollerMap.Store(c.fd,c)
}

func (p *Multiplexer) Poller_del(c *WsEnchentConn) {
	fdnum := int(c.fd)
	err:=unix.EpollCtl(p.epoller, unix.EPOLL_CTL_DEL, fdnum, nil)
	if err != nil {
		slog.Error("delete conn error", slog.String("err", err.Error()))
		return
	}
	delete(p.pollerMap, fdnum)
	// p.pollerMap.Delete(fdnum)
	slog.Info("delete conn", slog.Int("fd", fdnum))

}
func Poller_create() *Multiplexer {
	epoller, err := unix.EpollCreate1(0)
	if err != nil {
		slog.Error("create epoller error", slog.String("err", err.Error()))
		return nil
	}
	return &Multiplexer{
		epoller:   epoller,
		pollerMap: make(map[int]*WsEnchentConn),
		// pollerMap: &sync.Map{},
	}
}

func WsEnchent_create(ws *websocket.Conn) *WsEnchentConn {
	socketFile, err := ws.UnderlyingConn().(*net.TCPConn).File()
	if err != nil {
		slog.Error("get socket file error", slog.String("err", err.Error()))
		return nil
	}
	return &WsEnchentConn{
		Conn:      ws,
		
		fd:        int(socketFile.Fd()),
	}
}

func (c *WsEnchentConn) RegisterHandler(event eventif.WsEvent) {
	c.EventHandler = event
}

func (c *WsEnchentConn) PollerSelect(confg *Multiplexer) {
	c.Poller = confg
}
