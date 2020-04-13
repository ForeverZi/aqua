package wconn

import (
	"os"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
	"log"
)

var defaultOptions = []Option{
	BufferSize(1024, 1024),
	SkipAuth(),
	AutoIncUID(),
	ClientSendSize(20),
	EchoMsg(),
	Breaker(3, 2*time.Second),
	SetLogger(log.New(os.Stdout, "aqua", log.LstdFlags)),
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
		conf.handler = &EchoHanlder{}
	}
}

func CustomerMsgHandler(handler Handler) Option {
	return func(conf *HubConf) {
		conf.handler = handler
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
		conf.handler = protocol
	}
}

func SetLogger(logger Logger) Option {
	return func(conf *HubConf){
		conf.logger = logger
	}
}
