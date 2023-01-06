# Extensible lxkns

**lxkns** has several well-defined extension points for its CLI, supported
container engines, and (container) decorators. These extension points are
managed using the
[`github.com/thediveo/go-plugger/v3`](https://github.com/thediveo/go-plugger)
module.

## Plugin Groups

In `go-plugger` v3 parlance, a "plugin group" basically is a dedicated function
or interface type for which plugins then can register their plugin
implementations. The **lxkns** core can then call the registered plugin
functionalities in appropriate places.

### CLI

The following two plugin group types are defined for extending the CLI arguments
and handling of the **lxkns** CLI tools and service.

- `cliplugin.SetupCLI`: a `func(*cobra.Command)` that gets passed a cobra root
  `Command` in order to register CLI flags.

- `cliplugin.BeforeCommand`: a `func() error` that is run before the root
  command or a subcommand runs. This typically is used to validate CLI flags.

### Container Engine Watchers

- `engineplugin.NewWatchers`: a `func(*cobra.Command) ([]*NamedWatcher, error)`
  that returns a list of (named) watchers for tracking the containers of a
  single container engine.
  - it's perfectly okay to return a `nil` slice instead of any watchers when
    there is nothing to watch.
  - multiple watchers can be returned in situations where there is not (just) a
    single system container engine service, but potentially multiple (per user)
    container engine services -- such as in case of [Podman](https://podman.io)
    (but don't you call them _dæmons_!)

### Decorators

- `decorator.Decorate`: a `func(engines []*model.ContainerEngine, labels
  map[string]string)` that operates on the containers found and decorates them
  with additional information, such as composer project and Kubernetes pod
  groups.
