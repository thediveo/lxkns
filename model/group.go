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

package model

// Group groups a set of containers by a particular criterium as identified by
// the Type/Flavor of this group.
type Group struct {
	// Name of group of containers.
	Name string `json:"name"`
	// Type of container group in form of a unique identifier.
	Type string `json:"type"`
	// Optional flavor of container, or the same as the Type.
	Flavor string `json:"flavor"`
	// Containers in this group.
	Containers []*Container `json:"-"`
	// Labels store additional discovery-related group meta information.
	Labels Labels `json:"labels"`
}

// AddContainer adds the specified container to this group, updating also the
// container's group memberships.
func (g *Group) AddContainer(c *Container) {
	g.Containers = append(g.Containers, c)
	c.Groups = append(c.Groups, g)
}
