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

- `Type`: `"docker.com"`
- `Flavor`: `"docker.com"`

## Containerd

#### Engine

- `Type`: `"containerd.io"`

#### Containers

- `Name`: either only the container name when from the `"default"` containerd
  namespace, otherwise in the form `namespace/name`.
- `Type`: `"containerd.io"`
- `Flavor`: `"containerd.io"`

> [!ATTENTION] As the containerd namespace (not: Linux-kernel namespace) named
> `"moby"` is used by the Docker engine for its containers and the names(!) of
> Docker containers are not accessible via containerd this particular namespace
> is always ignored.
