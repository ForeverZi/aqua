package aqua

import (
	"log"
	"net/http"

	"github.com/ForeverZi/aqua/wconn"
)

type Server struct{}

func (s *Server) ListenAndServe(addr string) *http.Server {
	log.Printf("这是变更")
	hub := wconn.NewHub()
	go hub.Run()
	mux := http.NewServeMux()
	mux.Handle("/ws", hub)
	server := &http.Server{
		Handler: mux,
		Addr:    addr,
	}
	go func() {
		log.Printf("aqua listen on:%v", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("aqua server listen:%v\n", err)
		}
		log.Println("aqua server stopped...")
	}()
	return server
}
