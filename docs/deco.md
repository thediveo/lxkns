# Decorations

No, this isn't about the icing on your cake and we don't care if you try to have
and eat it at the same time. Instead, this is about useful information about
containers – and especially groups of containers – beyond the containers
themselves. Decoration cover common use cases so applications get relieved from
this kind of boring "data reconditioning".

## (Docker) Composer Projects

Containers belonging to a (Docker/nerdctl) composer project are decorated using
project-specific groups.

- `Group.Name`: composer project name as specified in a container's
  `com.docker.compose.project` label.
- `Group.Type`: `"com.docker.compose.project"`
- `Group.Flavor`: `"com.docker.compose.project"`

## Kubernetes Pods

Containers that are part of a Kubernetes pod are decorated using pod groups.
**lxkns** currently supports both dockershim and CRI-containerd.

- `Group.Name`: in the form of `namespace/podname`.
- `Group.Type`: `"io.kubernetes.pod"`
- `Group.Flavor`: `"io.kubernetes.pod"`

The dockershim decorator additionally ensures that the aforementioned container
labels are even present for dockershim-managed containers:

- `io.kubernetes.pod.uid`: pod UID.
- `lxkns/k8s/container/kind`: this label is present only if this container is a
  sandbox ("pause") container and always has an empty `""` value. This
  **lxkns**-specific decorator label allows applications to easily detect the
  sandbox containers of pods, regardless of how the containers are managed (CRI
  or dockershim).

## Decorator Plugins

Applications integrating **lxkns** can add their own decorators in order to
post-process containers after discovery and decorate them with groups, adapt
container flavors, et cetera. Decorators are registered using the simple
[thediveo/go-plugger](https://github.com/thediveo/go-plugger) plugin management
by calling `plugger.RegisterPlugin()` and specifying a
`decorator.Decorator`-compatible function.

```go
// Register this Decorator plugin.
func init() {
    plugger.RegisterPlugin(&plugger.PluginSpec{
        Name:  "mydecorator",
        Group: decorator.PluginGroup,
        Symbols: []plugger.Symbol{
            decorator.Decorate(Decorate),
        },
    })
}

// Decorate the discovered containers, where applicable...
func Decorate(engines []*model.ContainerEngine) {
    // ...
}
```

A decorator gets passed the list of container engines for which alive
(running/paused) containers have been discovered. The `model.ContainerEngine`s
then reference their managed `model.Container`s. For instance, decorators can create new `model.Group`s and add containers to these groups.

A good example is `decorator/compose/decorator.go`: this decorator looks for
(Docker) composer-related project labels. If found, it adds such containers into
their corresponding project groups.

```go
func Decorate(engines []*model.ContainerEngine) {
    for _, engine := range engines {
        projects := map[string]*model.Group{}
        for _, container := range engine.Containers {
            projectname, ok := container.Labels[ComposerProjectLabel]
            if !ok {
                continue
            }
            project, ok := projects[projectname]
            if !ok {
                project = &model.Group{
                    Name:   projectname,
                    Type:   ComposerGroupType,
                    Flavor: ComposerGroupType,
                }
                projects[projectname] = project
            }
            project.AddContainer(container)
        }
    }
}
```
