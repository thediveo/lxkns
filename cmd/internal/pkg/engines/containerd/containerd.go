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
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
	"github.com/thediveo/lxkns/cmd/internal/pkg/engines/engineplugin"
	"github.com/thediveo/whalewatcher/watcher/containerd"
)

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Group[engineplugin.NewWatchers]().Register(
		NewWatchers, plugger.WithPlugin("containerd"))
	plugger.Group[cliplugin.SetupCLI]().Register(
		SetupCLI, plugger.WithPlugin("containerd"))
}

// SetupCLI registers the Docker-engine specific CLI options.
func SetupCLI(cmd *cobra.Command) {
	cmd.PersistentFlags().String("containerd", "/run/containerd/containerd.sock",
		"containerd engine API socket path")
	cmd.PersistentFlags().Bool("nocontainerd", false, "do not consult a containerd engine")
}

// NewWatchers returns a moby engine watcher taking the supplied optional CLI flags
// into consideration. If this engine shouldn't be watched then it returns nil.
func NewWatchers(cmd *cobra.Command) ([]*engineplugin.NamedWatcher, error) {
	if nocontainerd, _ := cmd.PersistentFlags().GetBool("nocontainerd"); !nocontainerd {
		apipath, _ := cmd.PersistentFlags().GetString("containerd")
		w, err := containerd.New(apipath, nil)
		if err != nil {
			return nil, err
		}
		return []*engineplugin.NamedWatcher{
			{Watcher: w, Name: "containerd"},
		}, nil
	}
	return nil, nil
}
