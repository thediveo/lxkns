// Copyright 2021 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// +build linux

package model

// Container is a deliberately limited and simplified view on "alive" containers
// (which are containers with processes but not without). This is all we need in
// the context of Linux-kernel namespaces.
type Container interface {
	// Identifier of this container; depending on the particular container
	// engine this might be a unique container instance ID (for instance,
	// Docker).
	ID() string
	// Container name, which might be the same as the ID (for instance,
	// in case of containerd containers).
	Name() string
	// Type of container in form of a unique identifier, such as "docker.com",
	// "containerd.io", et cetera.
	Type() string
	// Optional flavor of container, or the same as the Type.
	Flavor() string
	// PID of the initial (or "ealdorman") container process. This is always
	// non-zero, as a Containerizer must never return any dead (non-alive)
	// containers.
	PID() PIDType
	// true, if the process(e)s inside this container have been either paused
	// and are in the process of pausing; otherwise false.
	Paused() bool
	// Meta data in form of labels assigned to this container.
	Labels() ContainerLabels

	// Managing container engine instance.
	Engine() ContainerEngine
	// Initial container process (ealdorman) details object.
	Process() *Process
}

// ContainerLabels are labels as key=value pairs assigned to a container. Both
// keys and values are strings.
type ContainerLabels map[string]string

// ContainerEngine represents a single specific instance of a container engine.
type ContainerEngine interface {
	// Container engine instance identifier/data, such as a UUID, et cetera.
	ID() string
	// Identifier of the type of container engine, such as "docker.com",
	// "containerd.io", et cetera.
	Type() string
	// List of alive containers managed by this container engine.
	Containers() []Container
	// Container engine API path (in initial mount namespace).
	API() string
}
