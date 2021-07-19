// Container engine support.

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

package engines

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
	"github.com/thediveo/go-plugger"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/containerd"
	"github.com/thediveo/whalewatcher/watcher/moby"
)

var engines []Engine

type Engine enumflag.Flag

const (
	NoEngine Engine = iota
	Docker
	Containerd
	All
)

var EngineModes = map[Engine][]string{
	NoEngine:   {"none"}, // pseudo engine name
	All:        {"all"},  // pseudo engine name
	Docker:     {"docker", "moby"},
	Containerd: {"containerd"},
}

// Containerizer returns a Containerizer watching the container engines
// specified on the CLI. Optionally waits for all container engines to come
// online.
func Containerizer(wait bool) (containerizer.Containerizer, error) {
	// TODO:
	if len(engines) == 0 {
		engines = []Engine{All}
	}
	if contains(NoEngine) {
		return nil, nil
	}

	watchers := []watcher.Watcher{}
	if contains(Docker) {
		w, err := moby.New("", nil)
		if err != nil {
			return nil, err
		}
		watchers = append(watchers, w)
	}
	if contains(Containerd) {
		w, err := containerd.New("", nil)
		if err != nil {
			return nil, err
		}
		watchers = append(watchers, w)
	}
	if len(watchers) == 0 {
		return nil, nil
	}
	// TODO: context
	cizer := whalefriend.New(context.Background(), watchers)
	for _, watcher := range watchers {
		<-watcher.Ready()
	}
	return cizer, nil
}

func contains(engine Engine) bool {
	for _, e := range engines {
		if e == engine || (engine != NoEngine && e == All) {
			return true
		}
	}
	return false
}

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.RegisterPlugin(&plugger.PluginSpec{
		Name:  "controlgroup",
		Group: "cli",
		Symbols: []plugger.Symbol{
			plugger.NamedSymbol{Name: "SetupCLI", Symbol: EngineSetupCLI},
		},
	})
}

func EngineSetupCLI(cmd *cobra.Command) {
	engines = []Engine{All}
	cmd.PersistentFlags().VarP(
		enumflag.NewSlice(&engines, "enginetype",
			EngineModes, enumflag.EnumCaseInsensitive),
		"engine", "e",
		"container engines to query; can be 'docker', 'containerd', 'none' or 'all'")
}
