package wconn

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

var (
	writeWait = 10 * time.Second

	readWait = 1200 * time.Second

	maxMessageSize int64 = 1024
)

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	id      int64
	send    chan []byte
	limiter *rate.Limiter
	MsgID   int64
	SendWnd []AckMsg
}

func (self *Client) SendHubCommand(command Command) error {
	return self.hub.SendCommand(command)
}

func (self *Client) Close() {
	self.OnUnregistered()
	self.conn.Close()
}

func (self *Client) Send(msg []byte) {
	self.send <- msg
}

func (self *Client) Ack(id int64) {
	for k, v := range self.SendWnd {
		if v.ID == id {
			v.Result <- true
			copy(self.SendWnd[k:], self.SendWnd[k+1:])
			self.SendWnd = self.SendWnd[:len(self.SendWnd)-1]
			return
		}
	}
}

func (self *Client) SendWithAck(msg AckMsg) (success bool) {
	sendmsg := func() bool {
		select {
		case <-msg.Ctx.Done():
			log.Printf("client.send阻塞\t%v:%v", "id", self.GetID())
			return false
		case self.send <- msg.Msg:
			return true
		}
	}
	if cap(self.send) == len(self.send) {
		success = sendmsg()
		self.SendWnd = append(self.SendWnd, msg)
	} else {
		self.SendWnd = append(self.SendWnd, msg)
		success = sendmsg()
	}
	return
}

func (self *Client) Write() {
	defer self.Close()
	for msg := range self.send {
		self.conn.SetWriteDeadline(time.Now().Add(writeWait))
		err := self.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			// 发送失败则注销此客户端
			self.Unregist()
			return
		}
	}
}

func (self *Client) Read() {
	defer self.Close()
	self.conn.SetReadLimit(maxMessageSize)
	for {
		self.conn.SetReadDeadline(time.Now().Add(readWait))
		_, msg, err := self.conn.ReadMessage()
		if err != nil {
			// 接受失败则注销此客户端
			self.Unregist()
			return
		}
		if self.limiter != nil && !self.limiter.Allow() {
			self.sendCloseMsg("busy")
			self.Unregist()
			return
		}
		if self.hub.conf.handleMsg != nil {
			err = self.hub.conf.handleMsg(self, msg)
			if err != nil {
				self.sendCloseMsg(err.Error())
				self.Unregist()
				return
			}
		}
	}
}

func (self *Client) sendCloseMsg(msg string) {
	self.conn.SetWriteDeadline(time.Now().Add(writeWait))
	self.conn.WriteMessage(websocket.CloseMessage, []byte(msg))
}

func (self *Client) Unregist() {
	self.hub.unregisterChan <- self
}

func (self *Client) Broadcast(msg []byte) {
	self.hub.Broadcast(msg)
}

func (self *Client) GetID() int64 {
	return self.id
}

func (self *Client) OnRegistered() (closed bool) {
	if self.hub.conf.onClientRegistered != nil {
		closed = self.hub.conf.onClientRegistered(self)
	}
	if closed {
		self.sendCloseMsg("regfail")
		self.Close()
		return
	}
	go self.Write()
	go self.Read()
	return
}

func (self *Client) OnUnregistered() {
	if self.hub.conf.onClientUnregistered != nil {
		self.hub.conf.onClientUnregistered(self)
	}
}
