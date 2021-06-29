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

// WhaleFriend is a containerizer internally backed by one or more
// Whalewatchers, that is, container watchers.
type WhaleFriend struct {
	watchers []watcher.Watcher
}

var _ containerizer.Containerizer = (*WhaleFriend)(nil)

// New returns a new containerizer using the specified set of container
// watchers. This also spins up the watchers to constantly watch in the
// background for any signs of container life and death.
func New(ctx context.Context, watchers []watcher.Watcher) containerizer.Containerizer {
	c := &WhaleFriend{
		watchers: watchers,
	}
	for _, watcher := range watchers {
		go watcher.Watch(ctx) // ...go on watch
	}
	return c
}

// watchersContainers returns the alive containers managed by the specified
// engine/watcher.
func (c *WhaleFriend) watchersContainers(ctx context.Context, engine watcher.Watcher) []model.Container {
	eng := containerizer.NewContainerEngine(
		engine.ID(ctx),
		engine.Type(),
		engine.API(),
		0 /* FIXME: unknown */)
	cntrs := []model.Container{}
	for _, projname := range append(engine.Portfolio().Names(), "") {
		project := engine.Portfolio().Project(projname)
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
				eng)
			cntrs = append(cntrs, cntr)
		}
	}
	return cntrs
}

// Containers returns the current container state of (alive) containers from all
// assigned whale watchers.
func (c *WhaleFriend) Containers(
	ctx context.Context, procs model.ProcessTable, pidmap model.PIDMapper,
) []model.Container {
	// Gather all alive containers known at this time to our whale watchers.
	containers := []model.Container{}
	for _, watcher := range c.watchers {
		containers = append(containers, c.watchersContainers(ctx, watcher)...)
	}
	return containers
}

// Close closes all watcher resources associated with this WhaleFriend.
func (c *WhaleFriend) Close() {
	for _, watcher := range c.watchers {
		watcher.Close()
	}
}
