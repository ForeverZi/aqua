package wconn

import (
	"context"
)

type AckMsg struct {
	ID  int64
	Ctx context.Context
	Msg []byte
	// 是否确认
	Result chan bool
}
