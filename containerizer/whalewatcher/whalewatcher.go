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

	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
)

// WhaleWatcher is a containerizer internally backed by one or more
// Whalewatchers, that is, container watchers.
type WhaleWatcher struct {
	engines []*ContainerEngine
}

var _ lxkns.Containerizer = (*WhaleWatcher)(nil)

// New returns a new containerizer using the specified set of container
// watchers. This also spins up the watchers to constantly watch in the
// background for any signs of container life and death.
func New(ctx context.Context, watchers []watcher.Watcher) lxkns.Containerizer {
	c := &WhaleWatcher{
		engines: make([]*ContainerEngine, len(watchers)),
	}
	for idx, watcher := range watchers {
		c.engines[idx] = NewContainerEngine(ctx, watcher)
	}
	return c
}

// Containerizer gets the current container state of (alive) containers from all
// assigned whale watchers and updates the discovery result data accordingly.
func (c *WhaleWatcher) Containerize(ctx context.Context, dr *lxkns.DiscoveryResult) {
	// Gather all alive containers known at this time to our whale watchers.
	containers := []model.Container{}
	for _, watcher := range c.engines {
		containers = append(containers, watcher.Containers()...)
	}
	// Update the discovery information with the container found and establish
	// the links between container and process information model objects.
	dr.Containers = containers
	for idx := range containers {
		// TODO: translate PID to initial PID namespace, if necessary.
		if proc, ok := dr.Processes[containers[idx].PID()]; ok {
			proc.Container = containers[idx]
			containers[idx].(*Container).process = proc
		}
	}
}

func (c *WhaleWatcher) Close() {
	for _, watcher := range c.engines {
		watcher.watcher.Close()
	}
}
