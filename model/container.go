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

//go:build linux

package model

// Containers is a slice of Container, offering some convenience functions, such
// as finding a container by name.
type Containers []*Container

// Container is a deliberately limited and simplified view on "alive" containers
// (where alive containers always have at least an initial process, so are never
// process-less). This is all we need in the context of Linux-kernel namespaces.
type Container struct {
	// Identifier of this container; depending on the particular container
	// engine this might be a unique container instance ID (for instance,
	// as in the case of a Docker-managed container).
	ID string `json:"id"`
	// Container name, which might be the same as the ID (for instance, in case
	// of containerd-managed containers), but might also be different (such as
	// in the case of Docker-managed containers).
	Name string `json:"name"`
	// Type of container in form of a unique identifier, such as "docker.com",
	// "containerd.io", et cetera.
	Type string `json:"type"`
	// Optional flavor of container, or the same as the Type.
	Flavor string `json:"flavor"`
	// PID of the initial (or "ealdorman") container process. This is always
	// non-zero, as Containerizers must never return any dead (non-alive)
	// containers. After finishing the discovery process this is the container's
	// PID in the initial PID namespace, even for containerized container
	// engines.
	PID PIDType `json:"pid"`
	// true, if the process(es) inside this container has (have) been either
	// paused and are in the process of pausing; otherwise false.
	Paused bool `json:"paused"`
	// Meta data in form of labels assigned to this container.
	Labels Labels `json:"labels"`

	// Group(s) this container belongs to.
	Groups []*Group `json:"-"`

	// Managing container engine instance.
	Engine *ContainerEngine `json:"-"`
	// Initial container process (ealdorman) details object.
	Process *Process `json:"-"`
}

// Labels are labels as key=value pairs assigned to a container. Both
// keys and values are strings.
type Labels map[string]string

// Group returns the group with specified Type, or nil.
func (c *Container) Group(Type string) *Group {
	for _, group := range c.Groups {
		if group.Type == Type {
			return group
		}
	}
	return nil
}

// FirstWithName returns the first container with the specified name, or nil if
// none could be found.
func (cs Containers) FirstWithName(name string) *Container {
	for _, c := range cs {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// WithEngineType returns only containers matching the specific engine type.
func (cs Containers) WithEngineType(enginetype string) (tcs Containers) {
	for _, c := range cs {
		if c.Engine.Type == enginetype {
			tcs = append(tcs, c)
		}
	}
	return
}

// FirstWithNameType returns the first container with the specified name and of
// the specified type (or flavor), or nil if none could be found.
func (cs Containers) FirstWithNameType(name string, typ string) *Container {
	for _, c := range cs {
		if c.Name == name && (c.Type == typ || c.Flavor == typ) {
			return c
		}
	}
	return nil
}
