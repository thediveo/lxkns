/*

lxkns is a micro-service for discovering Linux-kernel namespaces and related
information, such as namespace'd processes and the mapping of process
identifiers (PIDs) between hierarchical PID namespaces.

The lxkns service API definition can be found in api/openapi-spec/lxkns.yaml
inside the top-level project directory.

Usage

To use lxkns:

    lxkns [flag]

Flags

The following lxkns flags are available:

		--containerd string   containerd engine API socket path (default "/run/containerd/containerd.sock")
		--debug               enables debugging output
		--docker string       Docker engine API socket path (default "unix:///var/run/docker.sock")
	-h, --help                help for lxkns
		--http string         HTTP service address (default "[::]:5010")
		--initialcgroup       switches into initial cgroup namespace
		--nocontainerd        do not consult a containerd engine
		--nodocker            do not consult a Docker engine
		--noengines           do not consult any container engines
		--shutdown duration   graceful shutdown duration limit (default 15s)
		--silent              silences everything below the error level
	-v, --version             version for lxkns

*/

package main
