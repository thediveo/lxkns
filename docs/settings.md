# Settings

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
