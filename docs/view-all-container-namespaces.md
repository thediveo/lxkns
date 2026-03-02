# All Container (Namespaces) View

This view shows all discovered containers ➌ with their specific namespaces ➍.
The containers are organized by their managing container engines ➊. They are
additionally grouped by the lists of logical CPUs ➋ they are allowed to be
executed.

![view all containers](_images/lxkns-all-containers-view.png ':class=framedscreenshot')

For instance, "0-1" means that the container processes are allowed to execute on
any logical CPUs (this example is from a 2 CPU VM). In contrast, "1" allows the
processes of its container to execute only on the second logical CPU, but not on
the first one.

> [!RANT] Back in the early months of 2020 when lxkns started to take shape, the
> idea of running container engines inside containers was somewhat nerdy, but
> otherwise "_why should anyone do this?_". In 2026 we have [development
> containers](https://containers.dev/) and [Github
> codespaces](https://github.com/features/codespaces) and Docker-in-Docker is
> actually a well maintained sort-of default.
