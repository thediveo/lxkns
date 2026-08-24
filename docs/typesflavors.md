# Container Types/Flavors

**lxkns** assigns types to container engines and their managed containers,
allowing to work with different as well as multiple container engines. More to
the details, "containerizers" assign the container and engine types. And
decorators later might assign container *flavors* differing from the original
container *types*.

## Docker

#### Engine

- `Type`: `"docker.com"`

#### Containers

- "ordinary" containers -- in particular, just the genuine flavor.
  - `Type`: `"docker.com"`
  - `Flavor`: `"docker.com"`
- ["managed" plugin containers](https://docs.docker.com/engine/extend/) --
  flavored by the `dockerplugin` decorator.
  - `Type`: `"containerd.io"` – **NOTE** that Docker handles plugin containers
    as plain containerd containers. 
  - `Flavor`: `"plugin.docker.com"`
- [Siemens Industrial Edge](http://siemens.com/industrial-edge) App containers
  -- flavored by the `industrialedge` decorator.
  - `Type`: `"docker.com"`
  - `Flavor`: `"com.siemens.industrialedge.app"`

#### Groups

A group represents a single compose projects consisting of multiple "ordinary"
Docker containers.

- `Name`: name of Compose project.
- `Type`: `"com.docker.compose.project"`
- `Flavor`: `"com.docker.compose.project"`

## Containerd

#### Engine

- `Type`: `"containerd.io"`

#### Containers

- `Name`: either only the container name when from the `"default"` containerd
  namespace, otherwise in the form `namespace/name`.
- `Type`: `"containerd.io"`
- `Flavor`: `"containerd.io"`

> [!ATTENTION] The containerd namespaces (not: Linux-kernel namespace) named
> `"moby"` and `"plugins.moby"` are used by the Docker engine for its (plugin)
> containers. Now, the names(!) of Docker containers are not accessible via
> `containerd` so the `"moby"` namespace is always ignored by the
> `containerd`-specific discovery. However, the `"plugins.moby"` is handled as a
> containerd Container instead of a Docker container, and later post-processed
> by the `dockerplugin` decorator.

## CRI API

#### Engine

- `Type`: `"k8s.io/cri-api"`

#### Containers

- `Name`: of the container (without the namespace).
- `Type`: `"k8s.io/cri-api"`

#### Groups

Groups describe k8s pods:

- `Name`: of the k8s namespace.
- `Type`: `"io.kubernetes.pod"`
- `Flavor`: `"io.kubernetes.pod"`
