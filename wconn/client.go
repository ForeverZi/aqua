package wconn

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

var (
	writeWait = 10 * time.Second

	readWait = 1200 * time.Second

	maxMessageSize int64 = 1024
)

const (
	RGFAIL_CODE int = 1 + iota
	BUSY_CODE
	ERR_CODE
)

const (
	READ_CLOSED uint8 = 0x01 << iota
	WRITE_CLOSED
	UNREGISTERED
	SEND_CLOSE_MSG
)

type Client struct {
	mu      sync.Mutex
	hub     *Hub
	conn    *websocket.Conn
	id      int64
	send    chan []byte
	limiter *rate.Limiter
	MsgID   int64
	SendWnd []AckMsg
	cond    sync.Cond
	//标志: 读关闭(0x01),写关闭(0x02),反注册(0x04)
	flags    uint8
	closeMsg []byte
}

func (self *Client) SendHubCommand(command Command) error {
	return self.hub.SendCommand(command)
}

func (self *Client) Close() {
	self.OnUnregistered()
	close(self.send)
	self.mu.Lock()
	if self.cond.L == nil {
		self.cond.L = &self.mu
	}
	for !self.readyClose() {
		self.cond.Wait()
	}
	self.mu.Unlock()
	defer self.conn.Close()
	if len(self.closeMsg) > 0 {
		self.conn.SetWriteDeadline(time.Now().Add(writeWait))
		self.conn.WriteMessage(websocket.CloseMessage, self.closeMsg)
	}
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
	defer self.rwExitHandler(WRITE_CLOSED)()
	for msg := range self.send {
		self.conn.SetWriteDeadline(time.Now().Add(writeWait))
		err := self.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			// 发送失败则注销此客户端
			return
		}
	}
}

func (self *Client) Read() {
	defer self.Unregist()
	self.conn.SetReadLimit(maxMessageSize)
	for {
		self.conn.SetReadDeadline(time.Now().Add(readWait))
		_, msg, err := self.conn.ReadMessage()
		if err != nil {
			// 接受失败则注销此客户端
			return
		}
		if self.limiter != nil && !self.limiter.Allow() {
			self.sendCloseMsg(BUSY_CODE, "busy")
			return
		}
		if self.hub.conf.handleMsg != nil {
			err = self.hub.conf.handleMsg(self, msg)
			if err != nil {
				self.sendCloseMsg(ERR_CODE, err.Error())
				return
			}
		}
	}
}

func (self *Client) sendCloseMsg(code int, msg string) {
	self.mu.Lock()
	defer self.mu.Unlock()
	// 只接受第一次设置的关闭消息
	if self.flags&SEND_CLOSE_MSG > 0 {
		return
	}
	self.flags = self.flags | SEND_CLOSE_MSG
	self.closeMsg = websocket.FormatCloseMessage(code, msg)
}

func (self *Client) readyClose() bool {
	// readClose := self.flags & READ_CLOSED > 0
	writeClose := (self.flags & WRITE_CLOSED) > 0
	unregisted := (self.flags & UNREGISTERED) > 0
	return writeClose && unregisted
}

func (self *Client) rwExitHandler(setFlag uint8) func() {
	return func() {
		self.Unregist()
		self.mu.Lock()
		if self.cond.L == nil {
			self.cond.L = &self.mu
		}
		self.flags = self.flags | setFlag
		self.mu.Unlock()
		self.cond.Signal()
	}
}

func (self *Client) Unregist() {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.flags&UNREGISTERED == 0 {
		self.hub.unregisterChan <- self
		self.flags = self.flags | UNREGISTERED
	}
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
		self.sendCloseMsg(RGFAIL_CODE, "regfail")
		self.flags = self.flags | WRITE_CLOSED | READ_CLOSED
		self.Unregist()
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
