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
	"time"

	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
	"github.com/thediveo/lxkns/cmd/internal/pkg/engines/engineplugin"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/whalewatcher/watcher"

	_ "github.com/thediveo/lxkns/cmd/internal/pkg/engines/containerd" // pull in plugin
	_ "github.com/thediveo/lxkns/cmd/internal/pkg/engines/moby"       // pull in plugin
)

// Containerizer returns a Containerizer watching the container engines
// specified on the CLI. Optionally waits for all container engines to come
// online.
func Containerizer(ctx context.Context, cmd *cobra.Command, wait bool) (containerizer.Containerizer, error) {
	if ignoramus, _ := cmd.PersistentFlags().GetBool("noengines"); ignoramus {
		return nil, nil
	}
	watchers := []watcher.Watcher{}
	for _, plugf := range plugger.New(engineplugin.Group).Func("Watcher") {
		watcher, err := plugf.(engineplugin.NewWatcher)(cmd)
		if err != nil {
			return nil, err
		}
		if watcher != nil {
			watchers = append(watchers, watcher)
			log.Infof("synchronizing in background to %s engine, API %s",
				watcher.Name, watcher.API())
		}
	}
	if len(watchers) == 0 {
		return nil, nil
	}
	cizer := whalefriend.New(ctx, watchers)
	for _, w := range watchers {
		if wait {
			<-w.Ready()
			continue
		}
		go func(w *engineplugin.NamedWatcher) {
			select {
			case <-w.Ready():
				idctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				log.Infof("synchronized to %s engine with ID %s at API %s",
					w.Name, w.ID(idctx), w.API())
				cancel() // ensure to quickly release cancel
			case <-time.After(5 * time.Second):
				log.Warnf("%s engine still offline for API %s ... still trying in background",
					w.Name, w.API())
			}
		}(w.(*engineplugin.NamedWatcher))
	}
	return cizer, nil
}

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.RegisterPlugin(&plugger.PluginSpec{
		Name:  "engines",
		Group: cliplugin.Group,
		Symbols: []plugger.Symbol{
			plugger.NamedSymbol{Name: "SetupCLI", Symbol: EngineSetupCLI},
		},
	})
}

// EngineSetupCLI registers the engine-agnostic specific CLI options.
func EngineSetupCLI(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("noengines", false, "do not consult any container engines")
}
