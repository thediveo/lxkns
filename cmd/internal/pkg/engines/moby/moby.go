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

package moby

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v2"
	"github.com/thediveo/lxkns/cmd/internal/pkg/engines/engineplugin"
	"github.com/thediveo/whalewatcher/watcher/moby"
)

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Register(
		plugger.WithName("moby"),
		plugger.WithGroup(engineplugin.Group),
		plugger.WithNamedSymbol("Watchers", engineplugin.NewWatchers(Watchers)))
	plugger.Register(
		plugger.WithName("moby"),
		plugger.WithGroup("cli"),
		plugger.WithNamedSymbol("SetupCLI", SetupCLI))
}

// SetupCLI registers the Docker-engine specific CLI options.
func SetupCLI(cmd *cobra.Command) {
	cmd.PersistentFlags().String("docker", "unix:///var/run/docker.sock",
		"Docker engine API socket path")
	cmd.PersistentFlags().Bool("nodocker", false, "do not consult a Docker engine")
}

// Watchers returns a moby engine watcher taking the supplied optional CLI flags
// into consideration. If this engine shouldn't be watched then it returns nil.
func Watchers(cmd *cobra.Command) ([]*engineplugin.NamedWatcher, error) {
	if nodocker, _ := cmd.PersistentFlags().GetBool("nodocker"); !nodocker {
		apipath, _ := cmd.PersistentFlags().GetString("docker")
		w, err := moby.New(apipath, nil)
		if err != nil {
			return nil, err
		}
		return []*engineplugin.NamedWatcher{
			{Watcher: w, Name: "Docker"},
		}, nil
	}
	return nil, nil
}
