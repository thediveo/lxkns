# Linux kernel Namespaces
<img align="right" width="200" src="docs/_images/lxkns-gophers.jpeg">

…and Containers.

[![view manual](https://img.shields.io/badge/view-manual-blue)](https://thediveo.github.io/lxkns)
[![PkgGoDev](https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&labelColor=505050)](https://pkg.go.dev/github.com/thediveo/lxkns)
[![GitHub](https://img.shields.io/github/license/thediveo/lxkns)](https://img.shields.io/github/license/thediveo/lxkns)
![build and test](https://github.com/thediveo/lxkns/workflows/build%20and%20test/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/lxkns)](https://goreportcard.com/report/github.com/thediveo/lxkns)

`lxkns` is a Golang package for discovering Linux kernel namespaces as well as
mount points in mount namespaces. It then relates them to containers, where
possible. In every nook and cranny of your Linux hosts.

Discovery can be operated as a stand-alone REST service (both web UI and JSON
discovery results) or integrated into system diagnosis tools that need an
unobstructed view on Linux-kernel namespaces.

For mount namespaces, lxkns finds mount points even in process-less mount
namespaces (for instance, as utilized in ["snap"
technology](https://snapcraft.io/docs)). Our discovery engine even determines
the visibility of mount points, taking different forms of "overmounting" into
consideration.

Take a look at the comprehensive [user (and developer)
manual](https://thediveo.github.io/lxkns).

Or, watch the short overview video how to find your way around discovery web
frontend:

[![lxkns web
app](https://img.youtube.com/vi/4e6_jGLM9JA/0.jpg)](https://www.youtube.com/watch?v=4e6_jGLM9JA)

## ⚖️ Copyright and License

`lxkns` is Copyright 2020‒21 Harald Albrecht, and licensed under the Apache
License, Version 2.0.
