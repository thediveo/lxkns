# lxkns Service

## Container Deployment

Deployment features:

- **read-only:** the lxkns service can be used on a read-only container file
  system without any issues.

- **non-root:** the holy grail of container hardening … wait till you get to
  see our capabilities below.

- **unprivileged:** because that doesn't mean in-capable.

- **capabilities:** not much to see here, just…
  - `CAP_SYS_PTRACE` gives us access to the namespace information in the proc
    file system.
  - `CAP_SYS_ADMIN` and `CAP_SYS_ADMIN` allow us to switch (especially mount)
    namespaces in order to look into more places compared to standard discovery
    tools. Additionally, they allow us to switch the discovery service back into
    the initial cgroup namespace in order to discover correct cgroup hierarchy
    information. Similar, temporarily switching into the initial mount namespace
    allows us to correctly pick up the freezer ("fridge") states of processes,
    this works around having to either explicitly mount the host's cgroup into
    the container or to unprotect the container's system paths (which
    docker-compose yet does not support).
  - `CAP_DAC_READ_SEARCH` allows us to discover bind-mounted namespaces without
    interference by any in-descretionary excess control (DAC).

The convertainerized lxkns service correctly handles these pitfalls:

- **reading from other mount namespaces**: in order to discover mount points
  from a process-less bind-mounted mount namespace, lxkns forks itself and then
  re-executes the child in the mount namespace to read its procfs `mountinfo`
  from. The child here acts as the required procfs entry to be able to read the
  correct `mountinfo` at all. However, when containerized, lxkns runs in its own
  mount namespace, whereas the bindmount of the mount namespace will be in some
  other mount namespace, such as the host's initial mount namespace. In order to
  successfully reference the bindmount in the VFS, lxkns uses the Linux kernel's
  procfs wormholes: `/proc/[PID]/root/...`, see also
  [proc(5)](https://man7.org/linux/man-pages/man5/proc.5.html).

- **cgroup namespaced container**: during startup, lxkns detects when it has
  been placed into its own cgroup namespace ... as, for example, it is the case
  in newer Docker default installations on Linux base OS configurations
  especially with a cgroups v2 unified hierarchy. Without further measures, lxkns would be unable to discover the correct freezer states of processes. Thus, lxkns then switches itself out of its own cgroup namespace and back into the host's initial namespace, if possible. Please note that running lxkns in a non-initial namespace blocks correct discovery, not least process freezer state discovery.

## Behind Path-Rewriting Reverse Proxies

The web user interface of the lxkns service is a
[React-based](https://reactjs.org/) so-called [single-page
application](https://en.wikipedia.org/wiki/Single-page_application) (SPA). Now,
serving single-page applications using client-side HTML5 DOM routing behind
path-rewriting reverse proxies always is a challenge.

Of course, as long as you know the exact details where your containerized server
is going to be deployed you might statically set the final base path and build
your React SPA to exactly this configuration. However, if you want to make your
server image more versatile and know the reverse proxy (proxies) in front of
your server will cooperate by telling you the original URL as used by the
client, then things get more flexible.

For lxkns, we use this method: if there is a (rewriting) reverse proxy in front
of our service, it must pass a `X-Forwarded-Uri` HTTP request header with either
the full URL (URI) or at least the absolute path of the resource as originally
requested by a client. This allows our service to determine the "base" path by
comparing the path seen by our service versus the path seen by the first proxy.
This information is then used to dynamically rewrite the `<base href=""/>` from
`index.html` as needed.

In its `public/index.html`, lxkns sets `<base href="%PUBLIC_URL%/"/>` – **please
note the trailing slash!** This will work correctly for development as usual,
where the development server serves from the root.

For the production version we build with `PUBLIC_URL` set to "." (sic!) instead
to "/". This is not a mistake but ensures that all webpack-generated resources
are properly referenced **relative to the (dynamic) base URL**.

Of course, all other web app resources must be referenced using only relative
paths too, including the shortcut/favorite icon, et cetera. There must be no
`%PUBLIC_URL%/` anywhere, except for the `<base />` element.

And the lxkns (REST) API calls must also be relative, too.

In order to make HTML5 DOM routing properly work behind a path-rewriting reverse
proxy the lxkns SPA at runtime picks up its own `<base />` element path and then
passes that on to its DOM router; see `web/lxkns/src/utils/basename.ts` and
`web/lxkns/src/app/App.tsx` for how this is done in lxkns.
