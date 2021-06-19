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

package whalefriend

import (
	"context"

	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
)

// ContainerEngine implements the model.ContainerEngine perspective on container
// engine information.
type ContainerEngine struct {
	id      string          // container engine instance ID.
	watcher watcher.Watcher // container engine watcher.
	pid     model.PIDType   // engine PID, if known; otherwise, zero.
}

var _ (model.ContainerEngine) = (*ContainerEngine)(nil)
var _ (containerizer.Containerizer) = (*ContainerEngine)(nil)

// NewContainerEngine returns a new ContainerEngine connected to the specified
// container engine watcher, already up and running.
func NewContainerEngine(ctx context.Context, watcher watcher.Watcher, enginepid model.PIDType) *ContainerEngine {
	e := &ContainerEngine{
		watcher: watcher,
		id:      watcher.ID(ctx),
		pid:     enginepid,
	}
	go watcher.Watch(ctx) // ...fire up the watch engine
	return e
}

// Container engine instance identifier/data, such as a UUID, et cetera.
func (e *ContainerEngine) ID() string { return e.id }

// Identifier of the type of container engine, such as "docker.com",
// "containerd.io", et cetera.
func (e *ContainerEngine) Type() string { return e.watcher.Type() }

// Container engine API path (in initial mount namespace).
func (e *ContainerEngine) API() string { return e.watcher.API() }

// Container engine PID, if known. Otherwise, zero.
func (e *ContainerEngine) PID() model.PIDType { return e.pid }

// Containers returns the list of currently alive containers managed by this
// container engine.
func (e *ContainerEngine) Containers(
	ctx context.Context, _ model.ProcessTable, _ model.PIDMapper,
) []model.Container {
	cntrs := []model.Container{}
	for _, projname := range append(e.watcher.Portfolio().Names(), "") {
		project := e.watcher.Portfolio().Project(projname)
		if project == nil {
			continue
		}
		for _, container := range project.Containers() {
			cntr := containerizer.NewContainer(
				container.ID,
				container.Name,
				"", // default to Type derived from container engine.
				model.PIDType(container.PID),
				container.Paused,
				container.Labels,
				e)
			cntrs = append(cntrs, cntr)
		}
	}
	return cntrs
}

// Close releases the watcher and its resources.
func (e *ContainerEngine) Close() {
	e.watcher.Close()
}
