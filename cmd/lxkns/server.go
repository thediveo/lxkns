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
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/thediveo/lxkns/log"
)

// OriginalUrlHeader is the name of an optional HTTP request header passed to us
// by the first reverse proxy hit by a client's request. Its value allows us to
// determine the base path of our SPA as seen by the client. While
// X-Forwarded-Uri is generally barely documented it seems to be kind of a (oh,
// the irony) "well-known" header often used almost undetectedly.
//
// One place to spot it are Træfik's forward-request headers,
// https://doc.traefik.io/traefik/middlewares/forwardauth/#forward-request-headers.
const OriginalUrlHeader = "X-Forwarded-Uri"

// baseRe matches the base element in index.html in order to allow us to
// dynamically rewrite the base the SPA is served from. Please note that it
// doesn't make sense to use Go's templating here, as for development reasons
// the index.html must be perfectly usable without any Go templating at any
// time.
//
// Please note: "*?" instead of "*" ensures that our irregular expression
// doesn't get to greedy, gobbling much more than it should until the last(!)
// empty element.
var baseRe = regexp.MustCompile(`(<base href=").*?("\s*/>)`)

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

// httpError writes a normalized HTTP error message and HTTP status code given
// an error, not leaking any interesting internal server details from the
// original internal error.
func httpError(w http.ResponseWriter, e error) {
	if os.IsNotExist(e) {
		http.Error(w, "404 page not found", http.StatusNotFound)
	}
	if os.IsPermission(e) {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
	}
	http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
}

// ServeHTTP inspects the URL path to locate a file within the static dir on the
// SPA handler. If a file is found, it will be served. If not, the file located
// at the index path on the SPA handler will be served. This is suitable
// behavior for serving an SPA (single page application).
func (h appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the absolute (and cleaned) path to prevent directory traversal,
	// ensuring NOT to use the current working dir whichever it is.
	uriPath := path.Clean("/" + r.URL.Path)
	// Check whether a file exists at the given path; if it doesn't then serve
	// index.html from the "root" location of our static file assets instead.
	// Please note that we also consider all directories themselves to trigger
	// fallback to index.html.
	if info, err := os.Stat(path.Join(h.staticPath, uriPath)); err == nil && !info.IsDir() {
		// simply use http.FileServer to serve the existing static file; please
		// note that http.FileServer.ServeHTTP correctly sanitizes r.URL.Path
		// itself before trying to serve the filesystem resource, so it is kept
		// inside h.staticPath.
		log.Debugf("serving static resource %s", uriPath)
		http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
		return
	} else if err != nil && !os.IsNotExist(err) {
		// we got an error other than that the file doesn't exist when trying to
		// stat the file, so return a (sanitized) 500 internal server error and
		// be done with it.
		log.Debugf("no such static resource %s", uriPath)
		httpError(w, err)
		return
	}

	// determine the base path to the SPA as seen by clients. Here, we don't
	// want to rely on "magic" signatures in paths but instead rely on the first
	// reverse proxy correctly setting some HTTP request header. So, we're
	// relying on proxy magic instead... :o
	origUriPath := uriPath
	if clientUri, ok := r.Header[OriginalUrlHeader]; ok {
		if len(clientUri) > 0 && clientUri[0] != "" {
			// Again, sanitize whatever some self-acclaimed lobbxy, erm, proxy
			// sent us. For instance, the proxy's posh spell checker might have
			// flipped the URL to call the /reduce/tax API instead of
			// /reduce/VAT.
			if strings.HasPrefix(clientUri[0], "/") {
				origUriPath = path.Clean(clientUri[0])
			} else {
				// might be an URI, erm, URL, so try that; if that fails, we
				// just ignore it.
				if u, err := url.Parse(clientUri[0]); err == nil {
					origUriPath = path.Clean("/" + u.Path)
				}
			}
		}
	}
	var base string
	if !strings.HasSuffix(origUriPath, "/") {
		origUriPath += "/"
	}
	if strings.HasSuffix(origUriPath, uriPath) {
		base = origUriPath[:len(origUriPath)-len(uriPath)]
	} else {
		base = "" // fallback to root base in case the proxy passed us (ex-)PM nonsense.
	}
	if !strings.HasSuffix(base, "/") {
		// Ensure that the base path always ends with a "/", as otherwise
		// browsers will throw the specified path under the bus (erm, nevermind)
		// of a dirname() operation, clipping off the final element that once
		// was a proper directory name. Oh, well.
		base += "/"
	}
	// Sanitize the path further to not interfere with our regexp replacement
	// operation which uses "$1" and "$2" back references. So we simply
	// eliminate any "$" in the path, as this definitely is not VMS (shudder)
	// and we don't need no "$" in the SPA paths.
	base = strings.ReplaceAll(base, "$", "")
	log.Debugf("serving dynamic index.html with base %s at %s", base, uriPath)

	// Grab the index.html's contents into a string as we need to modify it
	// on-the-fly based on where we deem the base path to be. And finally serve
	// the updated contents.
	f, err := os.Open(filepath.Join(h.staticPath, h.indexPath))
	if err != nil {
		httpError(w, err)
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		httpError(w, err)
		return
	}
	indexhtmlcontents, err := ioutil.ReadAll(f) // retain pre-1.16 compatibility for now
	if err != nil {
		httpError(w, err)
		return
	}
	finalIndexhtml := baseRe.ReplaceAllString(string(indexhtmlcontents), "${1}"+base+"${2}")
	http.ServeContent(w, r, "index.html", fi.ModTime(), strings.NewReader(finalIndexhtml))
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
	r.PathPrefix("/api").HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) })

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
