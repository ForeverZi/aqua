package handler

import (
	"fmt"

	"github.com/ForeverZi/aqua/wconn"
	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type ActionCode int

const (
	ECHO ActionCode = iota
	ACTION_INVALID
)

var (
	ErrUnknowCode = fmt.Errorf("未识别的操作")
	ErrUnregisteredCode = fmt.Errorf("未注册的操作")
)

type ExMsg struct{
	Code	ActionCode
	Params	string
}

type ResponseFunc func(client *wconn.Client, msg ExMsg) error

func NewExHandler() *ExHandler{
	h := &ExHandler{
		m: make(map[ActionCode]ResponseFunc),
	}
	h.HandleFunc(ECHO, func(client *wconn.Client, msg ExMsg)error{
		client.Send([]byte(msg.Params))
		return nil
	})
	return h
}

type ExHandler struct{
	m 	map[ActionCode]ResponseFunc
}

func (exh *ExHandler) Response(client *wconn.Client, data []byte) error {
	var msg ExMsg
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}
	if msg.Code >= ACTION_INVALID {
		return ErrUnknowCode
	}
	f, ok := exh.m[msg.Code]
	if !ok {
		return ErrUnregisteredCode
	}
	return f(client, msg)
}

func (exh *ExHandler) HandleFunc(code ActionCode, f ResponseFunc){
	exh.m[code] = f
}