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

package containerdtest

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
)

// Pool represents a containerd client working on a specific (containerd)
// namespace.
type Pool struct {
	Namespace string
	Client    *containerd.Client
}

// Container represents a container belonging to a containerd Pool.
type Container struct {
	pool      *Pool
	Container containerd.Container
}

// NewPool creates a new containerd client that works in the specified
// containerd namespace and connects to the specified API path.
func NewPool(endpoint string, namespace string) (*Pool, error) {
	if endpoint == "" {
		endpoint = "/run/containerd/containerd.sock"
	}
	client, err := containerd.New(endpoint)
	if err != nil {
		return nil, err
	}
	pool := &Pool{
		Namespace: namespace,
		Client:    client,
	}
	return pool, nil
}

// context returns a context set to the pool's containerd namespace.
func (pool *Pool) context() context.Context {
	return namespaces.WithNamespace(context.Background(), pool.Namespace)
}

// Run creates a new container using the specified image ref and runs it using
// the specified args. If run is false, the container is then paused.
func (pool *Pool) Run(id string, ref string, run bool, args []string, opts ...containerd.NewContainerOpts) (*Container, error) {
	ctx := pool.context()

	// Pull image if not already in the content store.
	var image containerd.Image
	var err error
	if image, err = pool.Client.GetImage(ctx, ref); err != nil {
		image, err = pool.Client.Pull(ctx, ref, containerd.WithPullUnpack)
	}
	if err != nil {
		return nil, err
	}

	opts = append(opts[:],
		containerd.WithNewSnapshot(id+"-snapshot", image),
		containerd.WithNewSpec(
			oci.WithImageConfigArgs(image, args),
		),
	)
	cntr, err := pool.Client.NewContainer(ctx, id, opts...)
	if err != nil {
		return nil, err
	}
	c := &Container{pool: pool, Container: cntr}
	task, err := cntr.NewTask(ctx, cio.NewCreator())
	if err != nil {
		pool.Purge(c)
		return nil, err
	}
	if err := task.Start(ctx); err != nil {
		return nil, err
	}
	if !run {
		if err := task.Pause(ctx); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// Purge (delete) the specified container from the pool, including its task and
// snapshot, if any.
func (pool *Pool) Purge(c *Container) {
	ctx := pool.context()
	if task, err := c.Container.Task(ctx, nil); err == nil {
		_, _ = task.Delete(ctx, containerd.WithProcessKill)
	}
	_ = c.Container.Delete(ctx, containerd.WithSnapshotCleanup)
}

// PurgeID purges the specified container (including task and snapshot),
// identified by its ID/name.
func (pool *Pool) PurgeID(id string) {
	if c, err := pool.Client.LoadContainer(pool.context(), id); err == nil {
		pool.Purge(&Container{pool: pool, Container: c})
	}
}

// Status returns the container (task) status or containerd.Unknown.
func (c *Container) Status() containerd.ProcessStatus {
	if task, err := c.Container.Task(c.pool.context(), nil); err == nil {
		if status, err := task.Status(c.pool.context()); err == nil {
			return status.Status
		}
	}
	return containerd.Unknown
}
