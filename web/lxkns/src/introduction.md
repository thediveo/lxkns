<img src="public/lxkns192.png" style="height: 8ex; float: right;">

This UI style guide introduces and show-cases the UI components of the lxkns web
user interface. At the same time, it also serves as some weird kind of testbed
to see if and how our UI components (hopefully correctly) render.

For its global state management the lxkns UI react app doesn't use React Redux
but instead [jōtai](https://github.com/pmndrs/jotai). This does not only
encompass the settings but also the discovery interval and result. The rationale
here is that React Redux cannot handle the discovery data model with loops (that
is, references going forth and back between namespaces and processes).

- show/hide system processes.
- light/dark theme.
- discovery interval (value or null=off) and result.
