// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/thediveo/lxkns/log"
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
		log.Infof("starting lxkns server to serve at %s", listener.Addr().String())
		if err := server.Serve(listener); err != nil {
			log.Errorf("lxkns server error: %s", err.Error())
		}
	}()
	return listener.Addr(), nil
}

func stopServer(wait time.Duration) {
	once.Do(func() {
		if server != nil {
			log.Infof("gracefully shutting down lxkns server, waiting up to %s...",
				wait)
			ctx, cancel := context.WithTimeout(context.Background(), wait)
			defer cancel()
			_ = server.Shutdown(ctx)
			log.Infof("lxkns server stopped.")
		}
	})
}
