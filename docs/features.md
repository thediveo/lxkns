# Features

- finds all 8 types of currently defined Linux-kernel
  [namespaces](https://man7.org/linux/man-pages/man7/namespaces.7.html) even [in
  arcane places](discovery), such as bind mounts and open file descriptors.

- gives namespaces names (sic!).

- [discovers the the freezer state](cgroup) and (freezer) cgroup controller path
  information for the processes attached to namespaces (transparently supports
  both [v1
  freezers](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v1/freezer-subsystem.html#cgroup-freezer)
  als well as [v2
  core](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html#core-interface-files)).

- discovers mount points in mount namespaces and derives the mount point
  visibility and VFS path hierarchy. The visibility identifies overmounts, which
  can either appear higher up the VFS path hierarchy but also "in place".

- the Go API supports not only discovery, but also switching namespaces (OS
  thread switching).

- tested with Go 1.13-1.16.

- namespace discovery can be integrated into other applications or run as a
  containerized discovery backend service with REST API and web front-end.

- marshal and unmarshal discovery results to and from JSON – this is especially
  useful for separating the super-privileged scanner process from non-privileged
  frontends.

- web front-end of discovery service can be deployed behind path-rewriting
  reverse proxies without any recompilation or image rebuilding when the first
  rewriting reverse proxy adds `X-Forwarded-Uri` HTTP request headers.

- CLI tools for namespace discovery and analysis.
