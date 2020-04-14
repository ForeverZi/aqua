package handler

import (
	"github.com/ForeverZi/aqua/encoder"
	"github.com/ForeverZi/aqua/wconn"
)

type Mux interface {
	HandleFunc(code ActionCode, f ResponseFunc)
	Encoder() encoder.MsgProto
	wconn.Handler
}
