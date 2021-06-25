# Features

- finds all 8 types of currently defined Linux-kernel
  [namespaces](https://man7.org/linux/man-pages/man7/namespaces.7.html).

- gives namespaces names (sic!).

- the Go API supports not only discovery, but also switching namespaces (both OS
  thread switching as well as forked child re-execution).

- tested with Go 1.13-1.16.

- namespace discovery can be integrated into other applications or run as a
  containerized discovery backend service with REST API and web front-end.

- web front-end can be deployed behind path-rewriting reverse proxies without
  any recompilation or image rebuilding for reverse proxies sending
  `X-Forwarded-Uri` HTTP request headers.

- marshal and unmarshal discovery results to and from JSON – this is especially
  useful for separating the super-privileged scanner process from non-privileged
  frontends.

- CLI tools for namespace discovery and analysis.
