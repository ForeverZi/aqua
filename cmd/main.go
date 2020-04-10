package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ForeverZi/aqua"
)

func main() {
	s := &aqua.Server{}
	server := s.ListenAndServe(":8080")
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
	<-quitChan
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("关闭服务器错误:%v", err)
	} else {
		log.Printf("服务器已正常关闭")
	}
}
