package aqua

import (
	"net/http"

	"github.com/ForeverZi/aqua/encoder"
	"github.com/ForeverZi/aqua/handler"
	"github.com/ForeverZi/aqua/wconn"
	"github.com/ForeverZi/aqua/log"
)

type Server struct{}

func (s *Server) ListenAndServe(addr string) *http.Server {
	logger := log.New()
	hub := wconn.NewHub(wconn.CustomerMsgHandler(handler.NewExHandler(encoder.JSON)), 
		wconn.SetLogger(logger))
	mux := http.NewServeMux()
	mux.Handle("/ws", hub)
	server := &http.Server{
		Handler: mux,
		Addr:    addr,
	}
	go func() {
		logger.Println("aqua start", "listen:", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Println("aqua interrupt", "err", err)
		}
		logger.Println("aqua server stopped...")
	}()
	return server
}
