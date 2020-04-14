package aqua

import (
	"net/http"

	"github.com/ForeverZi/aqua/encoder"
	"github.com/ForeverZi/aqua/handler"
	"github.com/ForeverZi/aqua/log"
	"github.com/ForeverZi/aqua/wconn"
)

func NewServer() *Server {
	logger := log.New()
	mux := handler.NewExHandler(encoder.JSON)
	return &Server{
		Logger: logger,
		Mux:    mux,
	}
}

type Server struct {
	Logger wconn.Logger
	Mux    handler.Mux
}

func (s *Server) ListenAndServe(addr string) *http.Server {
	logger := s.Logger
	hub := wconn.NewHub(wconn.CustomerMsgHandler(s.Mux),
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

func (s *Server) HandleFunc(code handler.ActionCode, f handler.ResponseFunc){
	s.Mux.HandleFunc(code, f)
}

func (s *Server) Encoder() encoder.MsgProto {
	return s.Mux.Encoder()
}
