# Tinkering

We recommend tinkering with **lxkns** in a [Development
Container](https://containers.dev/). The repository contains multiple
configurations, for starter we recommend the "lxkns (fully privileged,
Docker-in-Docker)" configuration.

This will drop you after some time into your dedicated development container
that – more for fun than really necessary – its own Docker engine. The
development container is configured with `pid:host` and `cgroupns:host` so that
any **lxkns** tool process or service has full host view.

## Build and Deploy lxkns Service

```bash
make deploy
```

## Tinkering with the UI

First time after (re)building the dev container (you might need to confirm that
Corepack downloads a specific yarn version):

```bash
cd web/lxkns
yarn install
```

Then any time inside `web/lxkns`:

```bash
yarn dev
```

To tinker with the storybook, inside `web/lxkns`:

```bash
yarn storybook
```

## Note

It's funny to see how people really get happy when `--privileged` gets dropped,
yet `CRAP_SYS_ADMIN` and `CAP_SYS_PTRACE` doesn't ring any bells – when these
should ring for kingdom come. The development container actually indirectly uses
`--privileged`: this is activated by the Docker-in-Docker devcontainer feature.
