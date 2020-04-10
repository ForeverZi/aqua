package wconn

import (
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

var defaultOptions = []Option{
	BufferSize(1024, 1024),
	SkipAuth(),
	AutoIncUID(),
	ClientSendSize(20),
	EchoMsg(),
	Breaker(3, 2*time.Second),
}

type Option func(conf *HubConf)

func BufferSize(readBufferSize, writeBufferSize int) Option {
	return func(conf *HubConf) {
		conf.upgrader.ReadBufferSize = readBufferSize
		conf.upgrader.WriteBufferSize = writeBufferSize
	}
}

func CustomerAuth(auth func(*http.Request) bool) Option {
	return func(conf *HubConf) {
		conf.auth = auth
	}
}

func SkipAuth() Option {
	return func(conf *HubConf) {
		conf.auth = func(r *http.Request) bool {
			return true
		}
	}
}

func FixedUID(uid int64) Option {
	return func(conf *HubConf) {
		conf.getuid = func(r *http.Request) int64 {
			return uid
		}
	}
}

func CustomerUID(getuid func(*http.Request) int64) Option {
	return func(conf *HubConf) {
		conf.getuid = getuid
	}
}

func AutoIncUID() Option {
	return func(conf *HubConf) {
		var assigned int64
		conf.getuid = func(r *http.Request) int64 {
			return atomic.AddInt64(&assigned, 1)
		}
	}
}

func ClientSendSize(size int) Option {
	return func(conf *HubConf) {
		conf.clientSendSize = size
	}
}

func ClientAuth() Option {
	return func(conf *HubConf) {
		conf.auth = func(r *http.Request) bool {
			return strings.ToLower(strings.TrimSpace(r.Header.Get("X-Auth"))) == "pass"
		}
	}
}

func EchoMsg() Option {
	return func(conf *HubConf) {
		conf.handleMsg = func(client *Client, msg []byte) error {
			client.Send(msg)
			return nil
		}
	}
}

func CustomerMsgHandler(handle func(*Client, []byte) error) Option {
	return func(conf *HubConf) {
		conf.handleMsg = handle
	}
}

func Breaker(cap int, period time.Duration) Option {
	return func(conf *HubConf) {
		conf.breakerCap = cap
		conf.breakerPeriod = period
	}
}

func OnClientRegister(handle func(client *Client) bool) Option {
	return func(conf *HubConf) {
		conf.onClientRegistered = handle
	}
}

func OnClientUnregister(handle func(client *Client)) Option {
	return func(conf *HubConf) {
		conf.onClientUnregistered = handle
	}
}

func ProtocolOption(protocol Protocol) Option {
	return func(conf *HubConf) {
		conf.onClientUnregistered = protocol.OnClientUnregister
		conf.onClientRegistered = protocol.OnClientRegister
		conf.handleMsg = protocol.Response
	}
}
