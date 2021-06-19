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
	engines []*ContainerEngine
}

var _ containerizer.Containerizer = (*WhaleFriend)(nil)

// New returns a new containerizer using the specified set of container
// watchers. This also spins up the watchers to constantly watch in the
// background for any signs of container life and death.
func New(ctx context.Context, watchers []watcher.Watcher) containerizer.Containerizer {
	c := &WhaleFriend{
		engines: make([]*ContainerEngine, len(watchers)),
	}
	for idx, watcher := range watchers {
		c.engines[idx] = NewContainerEngine(ctx, watcher, 0)
	}
	return c
}

// Containers returns the current container state of (alive) containers from all
// assigned whale watchers.
func (c *WhaleFriend) Containers(
	ctx context.Context, procs model.ProcessTable, pidmap model.PIDMapper,
) []model.Container {
	// Gather all alive containers known at this time to our whale watchers.
	containers := []model.Container{}
	for _, watcher := range c.engines {
		containers = append(containers, watcher.Containers(ctx, procs, pidmap)...)
	}
	return containers
}

func (c *WhaleFriend) Close() {
	for _, watcher := range c.engines {
		watcher.Close()
	}
}
