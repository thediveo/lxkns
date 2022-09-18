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
// +build linux

package whalefriend

import (
	"context"

	"github.com/gammazero/workerpool"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
)

// WhaleFriend is a containerizer internally backed by one or more
// Whalewatchers, that is, [watcher.Watcher] container observers.
type WhaleFriend struct {
	watchers   []watcher.Watcher
	numworkers uint
	workers    workerpool.WorkerPool
}

var _ containerizer.Containerizer = (*WhaleFriend)(nil)

// Option represents a function able to set a particular WhaleFriend option
// state.
type Option func(c *WhaleFriend)

// New returns a new [containerizer.Containerizer] using the specified set of
// container watchers. This also spins up the watchers to constantly watch in
// the background for any signs of container life and death.
//
// The following options are supported:
//   - [WithWorkers] set the maximum of concurrent workers used when querying
//     the container workload.
func New(ctx context.Context, watchers []watcher.Watcher, opts ...Option) containerizer.Containerizer {
	c := &WhaleFriend{
		watchers: watchers,
	}
	for _, opt := range opts {
		opt(c)
	}

	// establish a worker pool for concurrently querying the container workloads
	// from multiple container engines.
	if c.numworkers < 1 {
		c.numworkers = 1
	}
	c.workers = *workerpool.New(int(c.numworkers))

	for _, w := range watchers {
		go func(w watcher.Watcher) { _ = w.Watch(ctx) }(w)
	}
	return c
}

// WithWorkers sets the maximum number of workers for querying workload
// containers from multiple container engines concurrently.
func WithWorkers(num uint) Option {
	return func(c *WhaleFriend) {
		c.numworkers = num
	}
}

// watchersContainers returns the alive [model.Container] objects managed by the
// specified engine/watcher. The containers returned are additionally linked to
// a unique [model.ContainerEngine] and these container engines are also aware
// of their containers.
func (c *WhaleFriend) watchersContainers(ctx context.Context, engine watcher.Watcher) []*model.Container {
	eng := &model.ContainerEngine{
		ID:      engine.ID(ctx),
		Type:    engine.Type(),
		Version: engine.Version(ctx),
		API:     engine.API(),
		PID:     model.PIDType(engine.PID()),
	}
	for _, projname := range append(engine.Portfolio().Names(), "") {
		project := engine.Portfolio().Project(projname)
		if project == nil {
			continue
		}
		for _, container := range project.Containers() {
			// Ouch! Make sure to clone the Labels map and not simply pass it
			// directly on to the lxkns container objects. Otherwise decorators
			// adding labels would modify the labels shared through the
			// underlying container label source. So, clone the labels
			// (top-level only) and then happy decorating.
			clonedLabels := model.Labels{}
			for k, v := range container.Labels {
				clonedLabels[k] = v
			}
			cntr := &model.Container{
				ID:     container.ID,
				Name:   container.Name,
				Type:   eng.Type,
				Flavor: eng.Type,
				PID:    model.PIDType(container.PID),
				Paused: container.Paused,
				Labels: clonedLabels,
				Engine: eng,
			}
			eng.AddContainer(cntr)
		}
	}
	return eng.Containers
}

// Containers returns the current container state of (alive) containers from all
// assigned whale watchers.
func (c *WhaleFriend) Containers(
	ctx context.Context, procs model.ProcessTable, pidmap model.PIDMapper,
) []*model.Container {
	// Gather all alive containers known at this time to our whale watchers. For
	// fun, we run the workload queries concurrently using a (limited) worker
	// pool. First, we submit query jobs for all engines that we watch...
	ch := make(chan []*model.Container)
	for _, watcher := range c.watchers {
		watcher := watcher // ...oh well...
		c.workers.Submit(func() {
			log.Debugf("querying workload from '%s'", watcher.ID(ctx))
			ch <- c.watchersContainers(ctx, watcher)
		})
	}
	// ...and then we're collecting the results while the workers from the pool
	// churn on the submitted jobs.
	containers := []*model.Container{}
	for w := 0; w < len(c.watchers); w++ {
		containers = append(containers, <-ch...)
	}
	return containers
}

// Close closes all watcher resources associated with this [WhaleFriend].
func (c *WhaleFriend) Close() {
	c.workers.Stop()
	for _, watcher := range c.watchers {
		watcher.Close()
	}
}
