// Copyright 2022 Harald Albrecht.
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

package dockerplugin

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops/mountineer"
	"github.com/thediveo/whalewatcher/watcher/containerd"
)

// DockerPluginFlavor is the Docker Plugin flavor of containerd containers.
const DockerPluginFlavor = "plugin.docker.com"

// DockerPluginNamespace is the containerd namespace for Docker plugin
// containers, including the trailing separator.
const DockerPluginNamespace = "plugins.moby/"

const enginebundlepathLabel = "com.docker/engine.bundle.path"

const managedPluginNameLabel = DockerPluginFlavor + "/name"

// Register this Decorator plugin.
func init() {
	plugger.Group[decorator.Decorate]().Register(
		Decorate, plugger.WithPlugin("dockerplugin"))
}

// Decorate decorates those discovered containerd containers that are actually
// Docker plugin containers.
func Decorate(engines []*model.ContainerEngine, labels map[string]string) {
	total := 0
	for _, engine := range engines {
		// If it "ain't no" containerd, skip it, as we're looking specifically
		// for containerd engines and their particular Docker plugin containers.
		if engine.Type != containerd.Type {
			continue
		}
		var mntee *mountineer.Mountineer // lazy initialization only when needed
		var mnterr error
		for _, cntr := range engine.Containers {
			if !strings.HasPrefix(cntr.Name, DockerPluginNamespace) {
				continue
			}
			log.Debugf("found managed Docker plugin container %s", cntr.Name)
			cntr.Flavor = DockerPluginFlavor
			bundlePath := cntr.Labels[enginebundlepathLabel]
			total++
			if bundlePath == "" {
				bundlePath = cntr.ID + ".sock"
			}
			if mntee == nil {
				// If we don't have a mountineer yet, but there was a previous
				// failed attempt with the current engine to access its mount
				// namespace then do not try again and simply carry on with the
				// next container.
				if mnterr != nil {
					continue
				}
				enginePID := engine.PID
				if enginePID == 0 {
					// If the engine PID isn't known then assumed the engine is
					// in the initial mount namespace and use PID 1 instead.
					enginePID = 1
				}
				mntee, mnterr = mountineer.New(model.NamespaceRef{fmt.Sprintf("/proc/%d/ns/mnt", enginePID)}, nil)
				if mnterr != nil {
					log.Errorf("dockerplugin decorator: cannot access mount namespace of engine with API %s, reason: %s",
						engine.API, mnterr.Error())
					continue
				}
				defer mntee.Close()
			}
			// Try to figure out the plugin's name from the socket name ... the
			// rationale is that we don't know the corresponding Docker daemon
			// instance and thus we lack the plugin information solely
			// maintained at the Docker level and not passed through to the
			// containerd layer.
			bundleItems, err := mntee.ReadDir("/run/docker/plugins/" + filepath.Base(bundlePath))
			if err != nil {
				continue
			}
			for _, item := range bundleItems {
				if item.Type()&fs.ModeSocket == 0 || !strings.HasSuffix(item.Name(), ".sock") {
					continue
				}
				pluginname := strings.TrimSuffix(item.Name(), ".sock")
				log.Debugf("managed Docker plugin %s is now known as %s", cntr.Name, pluginname)
				if pluginname == "" {
					break
				}
				cntr.Name = pluginname
				cntr.Labels[managedPluginNameLabel] = pluginname
				break
			}
		}
	}
	if total > 0 {
		log.Infof("discovered %d managed Docker plugins", total)
	}
}
