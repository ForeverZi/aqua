package wconn

import (
	"log"
)

type Handler interface {
	Response(client *Client, data []byte) error
}

type Protocol interface {
	OnClientRegister(client *Client) (closed bool)

	OnClientUnregister(client *Client)

	Handler
}

type Logger interface {
	Println(msg string, v ...interface{})
}

type DefaultLogger struct {
	logger *log.Logger
}

func (l *DefaultLogger) Println(msg string, v ...interface{}) {
	cv := []interface{}{msg}
	cv = append(cv, v...)
	l.logger.Println(cv...)
}
