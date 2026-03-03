# All Container (Namespaces) View

This view shows all discovered containers ➊ with their specific namespaces ➋.
The containers are organized by their managing container engines ➌.
Additionally, they are grouped by the sets of logical CPUs ➍ they are allowed to
be executed on.

![view all containers](_images/container-namespaces.png ':class=framedscreenshot')

For instance, "0-1" means that the container processes are allowed to execute on
any logical CPUs (this example is from a 2 CPU VM). In contrast, "1" allows the
processes of its container to execute _only_ on the _second_ logical CPU (#1),
but not on the first one (#0).

> [!RANT] Back in the early months of 2020 when lxkns started to take shape, the
> idea of running container engines inside containers was kind of nerdy, but
> mainly "_why should anyone do this?_". In 2026 we have [Development
> Containers](https://containers.dev/) and [Github
> codespaces](https://github.com/features/codespaces) – and Docker-in-Docker is
> actually a well maintained sort-of default.
