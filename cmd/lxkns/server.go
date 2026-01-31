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
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/spaserve"
)

var (
	once   sync.Once
	server *http.Server
)

// requestLogger is a middleware that closes the specified HTTP handler so that
// requests get logged at info level.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		slog.Info("http request", slog.String("method", req.Method), slog.String("uri", req.RequestURI))
		next.ServeHTTP(w, req)
	})
}

func startServer(address string, cizer containerizer.Containerizer) (net.Addr, error) {
	// Create the HTTP server listening transport...
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	// Finally create the request router and set the routes to the individual
	// handlers.
	r := mux.NewRouter()
	r.Use(requestLogger)
	r.HandleFunc("/api/namespaces", GetNamespacesHandler(cizer)).Methods("GET")
	r.HandleFunc("/api/processes", GetProcessesHandler).Methods("GET")
	r.HandleFunc("/api/pidmap", GetPIDMapHandler).Methods("GET")
	r.PathPrefix("/api").HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) })

	spa := spaserve.NewSPAHandler(os.DirFS("web/lxkns/build"), "index.html")
	r.PathPrefix("/").Handler(spa)

	server = &http.Server{
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		slog.Info("starting lxkns service", slog.String("addr", listener.Addr().String()))
		if err := server.Serve(listener); err != nil {
			slog.Error("lxkns service failed to start", slog.String("err", err.Error()))
		}
	}()
	return listener.Addr(), nil
}

func stopServer(wait time.Duration) {
	once.Do(func() {
		if server != nil {
			slog.Info("gracefully shutting down lxkns service",
				slog.Duration("grace", wait))
			ctx, cancel := context.WithTimeout(context.Background(), wait)
			defer cancel()
			_ = server.Shutdown(ctx)
			slog.Info("lxkns service stopped")
		}
	})
}
