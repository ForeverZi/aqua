package wconn

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

var (
	ErrUnknowCommand = fmt.Errorf("未定义的指令")
	ErrInvalidArgs   = fmt.Errorf("非法的指令参数")
	ErrInvalidChan   = fmt.Errorf("命令的结果通道必须未缓冲通道")
)

type HubConf struct {
	upgrader             websocket.Upgrader
	auth                 func(r *http.Request) bool
	getuid               func(r *http.Request) int64
	clientSendSize       int
	handler           	 Handler
	breakerCap           int
	breakerPeriod        time.Duration
	onClientRegistered   func(client *Client) bool
	onClientUnregistered func(client *Client)
}

type Hub struct {
	registerChan   chan *Client
	unregisterChan chan *Client
	broadcastChan  chan []byte
	commandChan    chan Command
	pool           map[int64]*Client
	conf           HubConf
}

type CommandOP int

const (
	ONLINE_COUNT_COMMAND CommandOP = iota
	GET_CLIENT_COMMAND
)

type Command struct {
	OP   CommandOP
	Args interface{}
	// 必须是缓存通道，否则可能会造成Hub主循环阻塞
	Result chan interface{}
}

func NewHub(options ...Option) *Hub {
	hub := Hub{
		registerChan:   make(chan *Client),
		unregisterChan: make(chan *Client),
		broadcastChan:  make(chan []byte, 256),
		commandChan:    make(chan Command, 20),
		pool:           make(map[int64]*Client),
	}
	for _, option := range defaultOptions {
		option(&hub.conf)
	}
	for _, option := range options {
		option(&hub.conf)
	}
	go hub.Run()
	return &hub
}

// 收到的Result首先需要检查类型是否是Error类型的，然后再处理对应指令期待的结构体类型
func (hub *Hub) SendCommand(command Command) error {
	if cap(command.Result) < 1 {
		return ErrInvalidChan
	}
	hub.commandChan <- command
	return nil
}

func (hub *Hub) handleCommand(command *Command) {
	switch command.OP {
	// 在接受到结果的时候首先要判断是否未error
	default:
		command.Result <- ErrUnknowCommand
	case ONLINE_COUNT_COMMAND:
		// 返回当前在线人数，int类型
		command.Result <- len(hub.pool)
	case GET_CLIENT_COMMAND:
		// 返回Error、nil或者指向client的指针
		if id, ok := command.Args.(int64); ok {
			command.Result <- hub.pool[id]
		} else {
			command.Result <- ErrInvalidArgs
		}
	}
}

func (hub *Hub) Run() {
	for {
		select {
		case command, ok := <-hub.commandChan:
			if !ok {
				return
			}
			hub.handleCommand(&command)
		case client, ok := <-hub.registerChan:
			if !ok {
				return
			}
			log.Printf("注册客户端:%v", client.id)
			hub.pool[client.id] = client
		case client, ok := <-hub.unregisterChan:
			if !ok {
				return
			}
			// 必须先检查当前的client是和需要注销的client是否一致
			currentClient := hub.pool[client.id]
			if currentClient == client {
				log.Printf("移除客户端:%v", client.id)
				delete(hub.pool, client.id)
			}
			go client.Close()
		case msg, ok := <-hub.broadcastChan:
			if !ok {
				return
			}
			// 这边拷贝一份是不是好很多
			log.Printf("广播消息:%v", string(msg))
			for _, client := range hub.pool {
				select {
				case client.send <- msg:
				default:
					// 如果无法立即发送到管道中，那么认为该客户端不可用
					delete(hub.pool, client.id)
					client.Close()
				}
			}
		}
	}
}

func (hub *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := hub.conf.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade fail:%v\n", err)
		return
	}
	if !hub.conf.auth(r) {
		log.Println("unauth conn!")
		conn.WriteControl(websocket.CloseMessage, []byte("authenticate required!"), time.Now().Add(10*time.Second))
		conn.Close()
		return
	}
	id := hub.conf.getuid(r)
	client := &Client{
		hub:     hub,
		conn:    conn,
		send:    make(chan []byte, hub.conf.clientSendSize),
		id:      id,
		limiter: rate.NewLimiter(rate.Every(hub.conf.breakerPeriod), hub.conf.breakerCap),
	}
	oldClient, ok := hub.pool[id]
	if ok {
		hub.unregisterChan <- oldClient
	}
	hub.registerChan <- client
	go client.OnRegistered()
}

func (hub *Hub) Broadcast(data []byte) {
	hub.broadcastChan <- data
}
