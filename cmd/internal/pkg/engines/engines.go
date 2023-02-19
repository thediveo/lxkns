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
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
	"github.com/thediveo/lxkns/cmd/internal/pkg/engines/engineplugin"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/whalewatcher/watcher"

	_ "github.com/thediveo/lxkns/cmd/internal/pkg/engines/containerd" // pull in plugin
	_ "github.com/thediveo/lxkns/cmd/internal/pkg/engines/moby"       // pull in plugin
)

// Names of the CLI flags defined and used in this package.
const (
	NoEnginesFlagName     = "noengines"
	KeepGoingFlagName     = "keep-going"
	EngineWorkersFlagName = "engine-workers"
)

// Containerizer returns a Containerizer watching the container engines
// specified on the CLI. Optionally waits for all container engines to come
// online.
func Containerizer(ctx context.Context, cmd *cobra.Command, wait bool) (containerizer.Containerizer, error) {
	if ignoramus, _ := cmd.PersistentFlags().GetBool(NoEnginesFlagName); ignoramus {
		return nil, nil
	}
	keepGoing, _ := cmd.PersistentFlags().GetBool(KeepGoingFlagName)
	watchers := []watcher.Watcher{}

	warned := false
	for _, newWatchers := range plugger.Group[engineplugin.NewWatchers]().PluginsSymbols() {
		log.Debugf("querying engine watcher plugin '%s'", newWatchers.Plugin)
		observers, err := newWatchers.S(cmd)
		if err != nil {
			log.Errorf("engine watcher plugin '%s' failure: %s", newWatchers.Plugin, err.Error())
			if !warned {
				warned = true
				log.Errorf("maybe you want to skip connecting to container engines using --%s?", NoEnginesFlagName)
			}
			if keepGoing {
				continue
			}
			return nil, err
		}
		if observers != nil {
			for _, observer := range observers {
				watchers = append(watchers, observer)
				log.Infof("synchronizing in background to %s engine, API %s",
					observer.Name, observer.API())
			}
		}
	}

	if len(watchers) == 0 {
		return nil, nil
	}

	numworkers, _ := cmd.PersistentFlags().GetUint(EngineWorkersFlagName)
	cizer := whalefriend.New(ctx, watchers, whalefriend.WithWorkers(numworkers))
	for _, w := range watchers {
		if wait {
			<-w.Ready()
			continue
		}
		go func(w *engineplugin.NamedWatcher) {
			idctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			engID := w.ID(idctx)
			log.Infof("%s engine with ID %s has version: %s",
				w.Name, engID, w.Version(idctx))
			cancel() // ensure to quickly release cancel
			// Oh, well: time.After is kind of hard to use without small leaks.
			// Now, a 5s timer will be GC'ed after 5s anyway, but let's do it
			// properly for once and all, to get the proper habit. For more
			// background information, please see, for instance:
			// https://www.arangodb.com/2020/09/a-story-of-a-memory-leak-in-go-how-to-properly-use-time-after/
			wecker := time.NewTimer(5 * time.Second)
			select {
			case <-w.Ready():
				if !wecker.Stop() { // drain the timer, if necessary.
					<-wecker.C
				}
				log.Infof("synchronized to %s engine with ID %s at API %s",
					w.Name, engID, w.API())
			case <-wecker.C:
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
	plugger.Group[cliplugin.SetupCLI]().Register(
		EngineSetupCLI, plugger.WithPlugin("engines"))
}

// EngineSetupCLI registers the engine-agnostic specific CLI options.
func EngineSetupCLI(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(NoEnginesFlagName, false, "do not consult any container engines")
	cmd.PersistentFlags().Bool(KeepGoingFlagName, true, "skip non-responsive container engines")
	cmd.PersistentFlags().Uint(EngineWorkersFlagName, 1, "maximum number of workers for container engine workload discovery")
}
