/*
dumpns runs a namespace (and process) discovery and then dumps the results as
JSON.

# Usage

To use dumpns:

	dumpns [flag]

For example, to view the discovery information as colored, pretty-printed JSON:

	dumpns -c | jq -C | less -SR

# Flags

The following dumpns flags are available:

	-c, --compact             compact instead of pretty-printed output
	    --containerd string   containerd engine API socket path (default "/run/containerd/containerd.sock")
	    --docker string       Docker engine API socket path (default "unix:///var/run/docker.sock")
	-h, --help                help for dumpns
	-i, --indent uint         use the given number of spaces (no more than 8) for indentation (default 2)
	    --nocontainerd        do not consult a containerd engine
	    --nodocker            do not consult a Docker engine
	    --noengines           do not consult any container engines
	-t, --tab                 use tabs for indentation instead of two spaces
	-v, --version             version for dumpns
*/
package main
