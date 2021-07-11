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

// ContainerEngine describes a single specific instance of a container engine.
type ContainerEngine struct {
	// Container engine instance identifier/data, such as a UUID, et cetera.
	ID string `json:"id"`
	// Identifier of the type of container engine, such as "docker.com",
	// "containerd.io", et cetera.
	Type string `json:"type"`
	// Container engine API path (in initial mount namespace).
	API string `json:"api"`
	// Container engine PID, if known. Otherwise, zero.
	PID PIDType `json:"pid"`

	// Containers discovered from this container engine.
	Containers []*Container `json:"-"`
}

// AddContainer adds a container to the list of discovered containers belonging
// to this particular engine instance. At the same time, it also links the
// container to this engine.
func (e *ContainerEngine) AddContainer(c *Container) {
	e.Containers = append(e.Containers, c)
	c.Engine = e
}
