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

package industrialedge

import (
	"strings"

	"github.com/thediveo/go-plugger"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/decorator/composer"
	"github.com/thediveo/lxkns/model"
)

// IndustrialEdgeAppFlavor identifies composer projects which are Industrial
// Edge apps.
const IndustrialEdgeAppFlavor = "com.siemens.industrialedge.app"

// EdgeRuntimeName is the name of the container housing the Siemens Industrial Edge
// runtime.
const EdgeRuntimeName = "edge-iot-core"

// FIXME:
const EdgeAppPrefix = "com_mwp_"

// Register this Decorator plugin.
func init() {
	plugger.RegisterPlugin(&plugger.PluginSpec{
		Name:      "industrialedge",
		Group:     decorator.PluginGroup,
		Placement: ">composer",
		Symbols: []plugger.Symbol{
			decorator.Decorate(Decorate),
		},
	})
}

// Decorate decorates the discovered Docker containers with Industrial Edge app
// project flavor, where applicable.
func Decorate(engines []*model.ContainerEngine) {
	for _, engine := range engines {
		isEdge := false
		for _, container := range engine.Containers {
			if container.Name == EdgeRuntimeName {
				isEdge = true
				// FIXME: decorate
			}
			break
		}
		if !isEdge {
			continue
		}
		// Decorate the IE app flavor for composer projects, where applicable.
		for _, container := range engine.Containers {
			isEdgeApp := false
			for labelkey := range container.Labels {
				if strings.HasPrefix(labelkey, EdgeAppPrefix) {
					isEdgeApp = true
					break
				}
			}
			if !isEdgeApp {
				continue
			}
			// FIXME: decorate
			if project := container.Group(composer.ComposerGroupType); project != nil {
				project.Flavor = IndustrialEdgeAppFlavor
			}
		}
	}
}
