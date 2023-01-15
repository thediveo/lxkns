# Docker Managed Plugins

> "Docker Engine’s plugin system allows you to install, start, stop, and remove
> plugins using Docker Engine." ([Docker
> documentation](https://docs.docker.com/engine/extend/))

Unfortunately, technical details are rather ... _scarce_.

In consequence, the following technical details base on our own analysis of the
Docker engine around version 20.10.

**Docker** handles its managed plugins as containerd containers in a separate
`"plugins.moby"` **containerd**(!) namespace. This way, plugin containers are
kept strictly separate from the ordinary **Docker** containers from the `"moby"`
**containerd** namespace.

Unfortunately, **Docker** (again) does not attach the plugin name to these
**containerd** containers -- this unfortunately mimics what we've already seen
for the **Docker** containers that are also "nameless" at the **containerd**
level.

Now, as Docker's managed plugins only have unique hex IDs at the containerd
level: how can we discover their plugin names? True, the Docker API has a
dedicated section for listing and inspecting plugins. Sadly, this plugin API
does **not** reveal the ID of the plugin container at the containerd level.
🤦‍♂️

The `dockerplugin` decorator of **lxkns** now pulls off an ugly hack, completely
ignoring the useless Docker plugin API: as it turns out, plugin containerd
containers have a label called `com.docker/engine.bundle.path` and this "bundle
path" actually points to the plugin's API socket. In case it isn't present, we
simply substitute the container's ID plus a `.sock` suffix.

Now where does this get us? Time to bring in the [mountineers](mountineers)! As
we know the PID and thus the mount namespace of the `containerd` engine process,
we then take a look inside(!) `/run/docker/plugins` to see if we can find there
a directory with the (basename) of the plugin's bundle path. And inside that
directory there's the plugin's API socket, properly named. At least, when using
Docker's plugin toolkit.

Yes, this is bad. Not looking at
[P.o.'d.man](http://thediveo.github.io/#/art/podman), it could be worse.