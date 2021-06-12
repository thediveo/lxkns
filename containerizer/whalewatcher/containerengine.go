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

package whalewatcher

import (
	"context"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
)

// ContainerEngine implements the model.ContainerEngine perspective on container
// engine information.
type ContainerEngine struct {
	id      string          // container engine instance ID
	typ     string          // container engine type ID
	watcher watcher.Watcher // container engine watcher
}

var _ (model.ContainerEngine) = (*ContainerEngine)(nil)

// NewContainerEngine returns a new ContainerEngine connected to the specified
// container engine watcher, already up and running.
func NewContainerEngine(ctx context.Context, watcher watcher.Watcher) *ContainerEngine {
	ce := &ContainerEngine{
		watcher: watcher,
		id:      watcher.ID(ctx),
		typ:     "", // FIXME:
	}
	go watcher.Watch(ctx) // ...fire up the watch engine
	return ce
}

func (e *ContainerEngine) ID() string { return e.id }

func (e *ContainerEngine) Type() string { return e.typ }

// Containers returns the list of currently alive containers managed by this
// container engine.
func (e *ContainerEngine) Containers() []model.Container {
	cntrs := []model.Container{}
	for _, projname := range append(e.watcher.Portfolio().Names(), "") {
		project := e.watcher.Portfolio().Project(projname)
		if project == nil {
			continue
		}
		for _, container := range project.Containers() {
			cntr := &Container{
				id:     container.ID,
				name:   container.Name,
				pid:    model.PIDType(container.PID),
				paused: container.Paused,
				labels: container.Labels,
				engine: e,
			}
			cntrs = append(cntrs, cntr)
		}
	}
	return cntrs
}

func (e *ContainerEngine) API() string { return "" } // FIXME:
