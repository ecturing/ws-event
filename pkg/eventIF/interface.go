package eventif

import "bytes"

type WsEvent interface {
	OnMessage(raw *bytes.Buffer)
	OnClose(raw *bytes.Buffer) error
	OnError(raw *bytes.Buffer, err error) //
}
