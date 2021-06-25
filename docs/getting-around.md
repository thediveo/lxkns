# Getting Around

After [deploying the lxkns service](getting-started), please navigate your web
browser to [http://localhost:5010](http://localhost:5010), where you should be
greeted by the web user interface of the lxkns service.

## Sidebar and Help

Clicking or tapping the sidebar symbol ❶ in the application bar opens the
sidebar. On touch devices, the sidebar can also be opened by swiping from the
left border to the right.

![application bar](_images/appbar.png ':class=framedscreenshot')

The application bar additionally offers these quick-access actions:

- ❷ collapses all tree levels, except for the root level as well as its
  immediate children.
- ❸ expands all tree levels.
- ❹ manual discovery and automatic refresh control. See also [discovery
  refresh](#discovery-refresh).

The sidebar can be closed either by click or tapping the close symbol ❶ or by
click or tapping outside the sidebar.

![sidebar](_images/sidebar.png ':class=framedscreenshot')

The sidebar also gives quick access to the integrated help ❷: multiple chapters
explain the displayed information in more detail.

## Discovery Refresh

The user interface defaults to manual refresh: clicking or tapping ❶ will start
a new discovery. If the discovery takes more time (such as on a system under
heavy load) or in case of a slow connection a progress indicator automatically
appears after approximately one second.

> [!NOTE] Discoveries cause some load on the target system, so please opt for
> either on-demand manual refreshes or choose a more relaxed refresh interval.

![refresh discovery](_images/refresh.png ':class=framedscreenshot')

Clicking or tapping ❷ opens a pop-up menu to either disable any automatic
refreshing or to choose from a set of available refresh intervals (500ms, 1s,
5s, 10s, 30s, 1min, 5min).

## All Namespaces View

The default view is named "all namespaces" and shows all discovered namespaces,
organized along the hierarchy of user namespaces ❶. This view reflects the
design of the Linux namespaces, where any namespace is always owned by a user
namespace. In case of user namespaces owning user namespaces this is also the
parent-child relationship.

![view all namespaces](_images/allview.png ':class=framedscreenshot')

lxkns shows for each user namespace ❶ the "most senior" process with its name
and PID ❷, as well as the user ID and user name to which this user namespace
belongs. The "most senior" process is also termed the "ealdorman"; it is the
topmost *and oldest* process in the process tree that is still attached to the
user namespace.

Now, there might be groups of processes attached to the user namespace with
different cgroup controllers ❸, ❹. These are additionally listed as to not miss
such different "tenants".

> [!NOTE] The init process with PID 1 is always shown first, while all other
> "tenants" are sorted by their process names.

Those namespaces ❸ created at system start are called "initial namespaces".
These are visually marked by dashed borders to make them easily spottable.

Other tenants ❹ might either use some or all of the existing namespaces ("shared
namespaces") or newly created namespaces instead. In case a namespace is
"shared" it is shown washed out. "Reused" namespaces can be hidden in the
[settings](#settings).

## Specific Namespaces

## Settings

The user interface and discovery display can be configured to some extent. The
settings are stored in the browser's web storage and is host-specific.

![settings](_images/settings.png ':class=framedscreenshot')

- **Theme**: switches between a light or dark theme, or considers the user's
  preference.
  
  > [!NOTE] Some desktop systems and browsers don't propagate any user
  > preferences to web applications, so this might default to a light theme
  > instead, regardless of the desktop theme set..

- **Show system processes**: shows or hides processes belonging to one of the
  following cgroups:
  - `/system.slice/*`
  - `/init.scope/*`
  - `/user.slice` (but not child cgroups)

  This setting defaults to "hide" in order to reduce the amount of process
  information.

- **Show shared non-user namespaces**: shows or hides namespaces that are used –
  "shared" – by multiple leader processes (and their subprocesses) with
  different cgroup controllers but the same owning user namespace. This setting
  defaults to showing all non-user namespaces of a leader process, regardless of
  whether a namespace is unique to that leader process, or not.

- **Expand newly discovered namespaces**: automatically expand all newly
  discovered namespaces, or leave them collapsed.
