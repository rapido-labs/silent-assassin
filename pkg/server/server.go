package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
)

type Server struct {
	apiServer *http.Server
	logger    logger.IZapLogger
}

//NewServer creates new server.
func NewServer(cp config.IProvider, zapLogger logger.IZapLogger) Server {

	router := mux.NewRouter()
	router.HandleFunc("/termination", ks.handlePreemption).Methods("POST")

	host := fmt.Sprintf("%s:%d", cp.GetString(config.ServerHost), cp.GetInt32(config.ServerPort))

	srv := &http.Server{
		Addr:    host,
		Handler: router,
	}

	return Server{
		apiServer: srv,
		logger:    zapLogger,
	}
}

//StartServer starts the server.
func (s Server) StartServer(ctx context.Context, wg *sync.WaitGroup) {
	s.logger.Info("Starting server")

	go s.listenServer(srv.apiServer)
	s.waitForShutdown(ctx)
}

func (s Server) listenServer() {
	if err := s.apiServer.ListenAndServe(); err != nil {
		s.logger.Error(fmt.Sprintf("Error starting server: %s", err.Error()))
		panic(err.Error())
	}
}

func (s Server) waitForShutdown(ctx context.Context, wg *sync.WaitGroup) {
	<-ctx.Done()
	s.apiServer.Shutdown(ctx)
	wg.Done()
	return
}
