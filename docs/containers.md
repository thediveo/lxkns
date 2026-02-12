# Containers

While namespace discovery truely is kernel-space territory it often is helpful
to correlate the discovered namespaces with user-space artefacts – and
containers in particular. Without doubt, containers are a dominant showcase of
namespace technology.

The container-related part of the **lxkns** discovery information model has on
offer:

- **containers** (as was to be expected),
- their managing container **engines**,
- container **groups**, such as [composer
  projects](https://github.com/compose-spec/compose-spec),
  [Kubernetes](https://kubernetes.io) pods, et cetera. A container can be a
  member of multiple groups at the same time, such as in the "have your cake"
  and "eat your cake" groups (but please don't expect things working in the
  virtual world to also work in reality).

**lxkns** uses these two architectural elements to discover containers and build
the container-related part of its information model:

- **containerizers** adapt lxkns to different container engines, with Docker and
  containerd support coming right out of the box. Multiple container engines can
  be handled simultaneously.

- **decorators** work on the meta information of the discovered containers and
  enhance ("decorate") the information model further with useful information
  about [composer projects](https://github.com/compose-spec/compose-spec),
  [Kubernetes](https://kubernetes.io) pods, and more.

## Containerizers

In **lxkns** parlance, a "containerizer" is tasked with querying one or more
container engines for their (alive) containers. So to say, to containerize
namespaces.

Right out of the box, **lxkns** handles the following container engines
supported by the [whalewatcher](https://github.com/thediveo/whalewatcher)
sibling project:

- Docker
- containerd
- CRI-O

> [!TIP] Applications can easily add their own containerizers (via so-called
> "watchers") to the lxkns service and the CLI tools, extending them via the
> `go-plugger` mechanism. Please see `cmd/internal/pkg/engines/moby/moby.go` for
> a good example.

## Decorators

**lxkns** features an extensible decorator plug-in infrastructure leveraging the
[go-plugger](https://github.com/thediveo/go-plugger) sibling project for Go
plugin management (but only using statically compiled-in plugins). Decorators
augment the found containers with additional information normally not directly
returned by container engine APIs as part of container inspection. For instance,
the composer detector (see next) creates composer project group elements based
on label annotations.

- [composer project](https://github.com/compose-spec/compose-spec) detection,
  based on `com.docker.composer.project` labels.

- [devcontainer](https://containers.dev/) and [Github
  codespace](https://github.com/features/codespaces) awareness, based on
  devcontainer-specific container labels and the metadata JSON information
  contained in some of these labels.

- Siemens [Industrial Edge](https://github.com/industrial-edge) app (and
  runtime) detection, which are a "flavor" of composer projects.

- [Kubernetes](https://kubernetes.io) pod detection:
  - [containerd CRI
    annotations](https://github.com/containerd/containerd/tree/main/pkg/cri),
    based on CRI-specific container labels.
  - [dockershim](https://github.com/kubernetes/kubernetes/tree/master/pkg/kubelet/dockershim),
    based on the esspecially encoded Docker container names. As the dockershim
    has been phased out of k8s a long time ago we'll drop this decorator in the
    not-to-far future.

> [!TIP] Applications using the `lxkns` module directly can seamlessly add their
> own decorators. They simply need to register them as "plugins" using the
> `go-plugger` mechanism. Please refer to the existing decorators in
> `decorator/` for details, such as `decorator/composer/decorator.go` as a good
> starter example.
