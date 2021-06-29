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

package containerizer

import (
	"github.com/thediveo/lxkns/model"
)

// ContainerEngine implements the model.ContainerEngine perspective on container
// engine information.
type ContainerEngine struct {
	id  string        // container engine instance ID.
	typ string        // container engine type identifier.
	api string        // API path.
	pid model.PIDType // engine PID, if known; otherwise, zero.
}

var _ (model.ContainerEngine) = (*ContainerEngine)(nil)

// NewContainerEngine returns a new ContainerEngine with the information specified.
func NewContainerEngine(id string, typ string, api string, pid model.PIDType) *ContainerEngine {
	return &ContainerEngine{
		id:  id,
		typ: typ,
		api: api,
		pid: pid,
	}
}

// Container engine instance identifier/data, such as a UUID, et cetera.
func (e *ContainerEngine) ID() string { return e.id }

// Identifier of the type of container engine, such as "docker.com",
// "containerd.io", et cetera.
func (e *ContainerEngine) Type() string { return e.typ }

// Container engine API path (in initial mount namespace).
func (e *ContainerEngine) API() string { return e.api }

// Container engine PID, if known. Otherwise, zero.
func (e *ContainerEngine) PID() model.PIDType { return e.pid }
