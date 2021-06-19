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

package containerizer

import (
	"github.com/thediveo/lxkns/model"
)

// Container implements the model.Container perspective on alive containers,
// providing storage for the required information.
//
// Please note that this implementation does not store the container Type, but
// instead returns the Type of the managing container engine.
type Container struct {
	id      string // container identifier.
	name    string // container name, might or might not differ from identifier.
	flavor  string // either sub-type or engine-independent container type.
	pid     model.PIDType
	paused  bool
	labels  model.ContainerLabels
	engine  model.ContainerEngine // references the managing container engine instance
	process *model.Process
}

// NewContainer returns a properly initialized container object implementing the
// model.Container interface. Please note that the associated process will only
// later be resolved and thus does not get specified here.
func NewContainer(id string, name string, flavor string, pid model.PIDType, paused bool, labels model.ContainerLabels, engine model.ContainerEngine) *Container {
	return &Container{
		id:     id,
		name:   name,
		flavor: flavor,
		pid:    pid,
		paused: paused,
		labels: labels,
		engine: engine,
	}
}

// Make sure the model.Container and model.ContainerFixer interfaces are fully
// implemented.
var _ model.Container = (*Container)(nil)
var _ model.ContainerFixer = (*Container)(nil)

// Identifier of this container; depending on the particular container
// engine this might be a unique container instance ID (for instance,
// as in the case of a Docker-managed container).
func (c *Container) ID() string { return c.id }

// Container name, which might be the same as the ID (for instance, in case
// of containerd-managed containers), but might also be different (such as
// in the case of Docker-managed containers).
func (c *Container) Name() string { return c.name }

// Type of container in form of a unique identifier, such as "docker.com",
// "containerd.io", et cetera. This implementation always derives the
// container Type from the managing container engine Type.
func (c *Container) Type() string { return c.engine.Type() }

// Optional flavor of container, or the same as the Type.
func (c *Container) Flavor() string {
	if c.flavor != "" {
		return c.flavor
	}
	return c.engine.Type()
}

// PID of the initial (or "ealdorman") container process. This is always
// non-zero, as a Containerizer must never return any dead (non-alive)
// containers. After finishing the discovery process this is the container's
// PID in the initial PID namespace, even for containerized container
// engines.
func (c *Container) PID() model.PIDType { return c.pid }

// true, if the process(es) inside this container has (have) been either
// paused and are in the process of pausing; otherwise false.
func (c *Container) Paused() bool { return c.paused }

// Meta data in form of labels assigned to this container.
func (c *Container) Labels() model.ContainerLabels { return c.labels }

// Managing container engine instance.
func (c *Container) Engine() model.ContainerEngine { return c.engine }

// Initial container process (ealdorman) details object.
func (c *Container) Process() *model.Process { return c.process }

// SetTranslatedPID sets a container's initial process PID as seen from the
// initial PID namespace.
func (c *Container) SetTranslatedPID(pid model.PIDType) { c.pid = pid }

// SetProcess sets the process (proxy) object corresponding with the
// containe3r's initial process PID.
func (c *Container) SetProcess(proc *model.Process) { c.process = proc }
