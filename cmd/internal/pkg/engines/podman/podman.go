//go:build podman

// Container engine support.

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

package podman

import (
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
	"github.com/thediveo/lxkns/cmd/internal/pkg/engines/engineplugin"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/sealwatcher/v2"
)

// Names of the CLI flags defined and used in this package.
const (
	PodmanFlagName     = "podman"
	UserPodmenFlagName = "user-podmen"
)

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Group[engineplugin.NewWatchers]().Register(
		NewWatchers, plugger.WithPlugin("podman"))
	plugger.Group[cliplugin.SetupCLI]().Register(
		SetupCLI, plugger.WithPlugin("podman"))
}

// SetupCLI registers the Podman-engine specific CLI options.
func SetupCLI(cmd *cobra.Command) {
	cmd.PersistentFlags().String(PodmanFlagName, "",
		"Podman service API socket path")
	cmd.PersistentFlags().Lookup(PodmanFlagName).NoOptDefVal = "unix:///run/podman/podman.sock"

	cmd.PersistentFlags().String(UserPodmenFlagName, "", "discover user podman services inside runtime directory")
	cmd.PersistentFlags().Lookup(UserPodmenFlagName).NoOptDefVal = "/run/user"
}

// NewWatchers returns a Podman engine watcher(s) taking the supplied optional CLI
// flags into consideration. If this engine shouldn't be watched then it returns
// nil.
func NewWatchers(cmd *cobra.Command) ([]*engineplugin.NamedWatcher, error) {
	watchers, err := systemPodmanWatcher(cmd)
	if err != nil {
		return nil, err
	}
	morewatchers, err := userPodmanWatchers(cmd)
	if err != nil {
		return nil, err
	}
	return append(watchers, morewatchers...), nil
}

// systemPodmanWatcher returns a Podman watcher for the system Podman service,
// if the "--podman" flag has been specified (with either the default API path,
// or an explicitly specified different API path).
func systemPodmanWatcher(cmd *cobra.Command) ([]*engineplugin.NamedWatcher, error) {
	apipath, _ := cmd.PersistentFlags().GetString(PodmanFlagName)
	if apipath == "" {
		return nil, nil
	}
	w, err := sealwatcher.New(apipath, nil)
	if err != nil {
		return nil, err
	}
	return []*engineplugin.NamedWatcher{
		{Watcher: w, Name: "Podman"},
	}, nil
}

// userPodmanWatchers scans inside the runtime directory specified by the
// "--user-podmen" flag for user-specific Podman service API endpoints,
// returning watchers for the ones found.
func userPodmanWatchers(cmd *cobra.Command) ([]*engineplugin.NamedWatcher, error) {
	usersRuntimeDir, _ := cmd.PersistentFlags().GetString(UserPodmenFlagName)
	if usersRuntimeDir == "" {
		return nil, nil
	}

	log.Debugf("scanning for user podman service endpoints inside %s", usersRuntimeDir)
	watchers := []*engineplugin.NamedWatcher{}
	usersRuntimeDirs, err := os.ReadDir(usersRuntimeDir)
	if err != nil {
		return nil, nil
	}
	for _, userRuntimeDir := range usersRuntimeDirs {
		if !userRuntimeDir.IsDir() {
			continue
		}
		user := userRuntimeDir.Name()
		apipath := path.Join(usersRuntimeDir, user, "podman/podman.sock")
		info, err := os.Stat(apipath)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSocket == 0 {
			continue
		}
		log.Debugf("found podman service endpoint for user %s", user)
		watcher, err := sealwatcher.New("unix://"+apipath, nil)
		if err != nil {
			log.Warnf("error accessing user %s podman service: %s", user, err)
			continue
		}
		watchers = append(watchers, &engineplugin.NamedWatcher{
			Watcher: watcher,
			Name:    "Podman-user-" + user,
		})
	}
	return watchers, nil
}
