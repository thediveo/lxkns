# Extensible lxkns

## Plugin Groups

### CLI

The plugin group named `"cli"` supports the following two symbols.

- `SetupCLI`: a `func(*cobra.Command)` that gets passed a cobra root `Command`
  in order to register CLI flags.

- `BeforeRun`: a `func() error` that is run before the root command or a
  subcommand runs. This typically is used to validate CLI flags.

### Container Engine Watchers

The plugin group named `"engines"` supports the following symbols.

- `engineplugin.NewWatcher`: a `func(*cobra.Command) (*NamedWatcher, error)`
  that returns a (named) watcher for tracking the containers of a single
  container engine.
  - it's okay to return `nil` instead of a watcher when there is nothing to
    watch.
