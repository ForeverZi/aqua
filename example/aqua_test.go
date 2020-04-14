package example

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"fmt"

	"github.com/ForeverZi/aqua"
	"github.com/ForeverZi/aqua/wconn"
	"github.com/ForeverZi/aqua/handler"
)

type ReqParams struct{
	X,Y	int
}

// 这是一个aqua服务器运行示例
// 在浏览器控制台使用:
// ws1 = new WebSocket("ws://localhost:8080/ws");
// ws1.onmessage = (evt)=>console.log("received:", evt.data);
// ws1.onclose = (evt)=>console.log("ws closed", evt);
// ws1.send(JSON.stringify({Code:1,Params:JSON.stringify({X:1, Y:2})}))
// 返回：received: {"Code":1,"Params":"3"}
func Example() {
	s := aqua.NewServer()
	logger := s.Logger
	// 注册处理函数
	s.HandleFunc(1, func(client *wconn.Client, msg handler.ExMsg)error{
		var param ReqParams
		err := s.Encoder().Unmarshal([]byte(msg.Params), &param)
		if err != nil {
			return err
		}
		msg.Params = fmt.Sprint(param.X+param.Y)
		resp, err := s.Encoder().Marshal(msg)
		if err != nil {
			return err
		}
		client.Send(resp)
		return nil
	})
	server := s.ListenAndServe(":8080")
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
	<-quitChan
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Println("关闭服务器错误", "err", err)
	} else {
		logger.Println("服务器已正常关闭")
	}
}
