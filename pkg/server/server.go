package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/killer"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
)

type Server struct {
	apiServer *http.Server
	logger    logger.IZapLogger
	killer    killer.KillerService
	cp        config.IProvider
}

//NewHttpServer creates new server.
func NewHttpServer(cp config.IProvider, zapLogger logger.IZapLogger, ks killer.KillerService) *Server {

	host := fmt.Sprintf("%s:%d", cp.GetString(config.ServerHost), cp.GetInt32(config.ServerPort))

	srv := &http.Server{
		Addr: host,
	}

	return &Server{
		apiServer: srv,
		logger:    zapLogger,
		killer:    ks,
		cp:        cp,
	}
}

func (s *Server) Start(ctx context.Context, wg *sync.WaitGroup) {
	s.logger.Info("Starting server")
	s.setRoutes()
	go s.listenServer()
	s.waitForShutdown(ctx, wg)
}

func (s *Server) setRoutes() {
	router := mux.NewRouter()
	router.HandleFunc("/termination", s.handleTermination).Methods(http.MethodPost)
	s.apiServer.Handler = router
}

func (s *Server) listenServer() {
	if err := s.apiServer.ListenAndServe(); err != nil {
		s.logger.Error(fmt.Sprintf("Error starting server: %s", err.Error()))
		panic(err.Error())
	}
	fmt.Println(s.apiServer.Handler)
}

func (s *Server) waitForShutdown(ctx context.Context, wg *sync.WaitGroup) {
	<-ctx.Done()
	s.apiServer.Shutdown(ctx)
	wg.Done()
	return
}
