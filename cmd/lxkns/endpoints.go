package main

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/api/types"
	"github.com/thediveo/lxkns/species"
)

func GetNamespacesHandler(w http.ResponseWriter, req *http.Request) {
	allns := lxkns.Discover(lxkns.FullDiscovery)
	// Note bene: set header before writing the header with the status code;
	// actually makes sense, innit?
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(
		types.NewDiscoveryResult(types.WithResult(allns))) // ...brackets galore!!!
	if err != nil {
		log.Errorf("namespaces discovery error: %s", err.Error())
	}
}

func GetProcessesHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
}

func GetPIDMapHandler(w http.ResponseWriter, req *http.Request) {
	opts := lxkns.FullDiscovery
	opts.NamespaceTypes = species.CLONE_NEWPID
	pidmap := lxkns.NewPIDMap(lxkns.Discover(opts))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(
		types.NewPIDMap(types.WithPIDMap(pidmap)))
	if err != nil {
		log.Errorf("namespaces discovery error: %s", err.Error())
	}
}
