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
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/thediveo/lxkns/log"
)

var (
	once   sync.Once
	server *http.Server
)

// appHandler implements the http.Handler interface, so we can use it to respond
// to HTTP requests. The path to the static directory and path to the index file
// within that static directory are used to serve the SPA in the given static
// directory.
type appHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir on the
// SPA handler. If a file is found, it will be served. If not, the file located
// at the index path on the SPA handler will be served. This is suitable
// behavior for serving an SPA (single page application).
func (h appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Infof("http %s %s", req.Method, req.RequestURI)
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

	spa := appHandler{staticPath: "web/lxkns/build", indexPath: "index.html"}
	r.PathPrefix("/").Handler(spa)

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
