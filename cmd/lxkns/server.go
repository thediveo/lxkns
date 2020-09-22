package main

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	once   sync.Once
	server *http.Server
)

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Infof("%s %s", req.Method, req.RequestURI)
		next.ServeHTTP(w, req)
	})
}

func startServer(address string) (net.Addr, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()
	r.Use(requestLogger)
	r.HandleFunc("/api/namespaces", GetNamespacesHandler).Methods("GET")
	r.HandleFunc("/api/processes", GetProcessesHandler).Methods("GET")
	r.HandleFunc("/api/pidmap", GetPIDMapHandler).Methods("GET")

	server = &http.Server{Handler: r}
	go func() {
		log.Info("starting lxkns server...")
		if err := server.Serve(listener); err != nil {
			log.Errorf("lxkns server error: %s", err.Error())
		}
	}()
	return listener.Addr(), nil
}

func stopServer(wait time.Duration) {
	once.Do(func() {
		if server != nil {
			log.Info("gracefully shutting down lxkns server...")
			ctx, cancel := context.WithTimeout(context.Background(), wait)
			defer cancel()
			_ = server.Shutdown(ctx)
			log.Info("lxkns server stopped.")
		}
	})
}
