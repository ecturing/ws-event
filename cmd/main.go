package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"ws-event/internal/dispatcher"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	dp = dispatcher.Default()
)

func main() {
	http.HandleFunc("/ws", wsServer)
	dispatcher.UseDispatcher(dp)
	slog.Info("start ws server,Listen on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		slog.Error("ListenAndServe:", slog.String("err", err.Error()))
	}
}

func wsServer(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("upgrade:", slog.String("err", err.Error()))
	}
	slog.Info("one ws connected", slog.String("addr:", conn.RemoteAddr().String()))
	dp.RegisterConn(conn, new(wsHandler))
}

type wsHandler struct {}

func (w wsHandler) OnClose(raw *bytes.Buffer) error {
	return nil
}

func (w wsHandler) OnMessage(raw *bytes.Buffer) {
	
	slog.Info("ws message:", slog.String("message", raw.String()))
}
func (w wsHandler) OnError(raw *bytes.Buffer, err error) {
	slog.Error("ws error:", slog.String("err", err.Error()))
}
