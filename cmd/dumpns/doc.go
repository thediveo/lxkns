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
	-h, --help                help for dumpns
	-i, --indent uint         use the given number of spaces (no more than 8) for indentation (default 2)
	-t, --tab                 use tabs for indentation instead of two spaces
	-v, --version             version for dumpns
	    --wait duration   max duration to wait for container engine workload synchronization (default 3s)
*/
package main
