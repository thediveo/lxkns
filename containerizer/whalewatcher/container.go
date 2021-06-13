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
	"github.com/thediveo/lxkns/model"
)

// Container implements the model.Container perspective on alive containers,
// providing storage for the required information.
type Container struct {
	id   string
	name string
	// TODO: flavor?
	pid     model.PIDType
	paused  bool
	labels  model.ContainerLabels
	engine  *ContainerEngine // references the managing container engine instance
	process *model.Process
}

// Make sure the model.Container interface is fully implemented.
var _ model.Container = (*Container)(nil)

func (c *Container) ID() string                    { return c.id }
func (c *Container) Name() string                  { return c.name }
func (c *Container) Type() string                  { return c.engine.Type() }
func (c *Container) Flavor() string                { return "" }
func (c *Container) PID() model.PIDType            { return c.pid }
func (c *Container) Paused() bool                  { return c.paused }
func (c *Container) Labels() model.ContainerLabels { return c.labels }
func (c *Container) Engine() model.ContainerEngine { return c.engine }
func (c *Container) Process() *model.Process       { return c.process }
