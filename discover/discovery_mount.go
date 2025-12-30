// Discovers mount paths with mount points in mount namespaces. This discovery
// requires an explicit opt-in. It needs to be run after the mount namespaces
// have been discovered. In consequence, it cannot be run without mount
// namespace discovery.

// Copyright 2021 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux

package discover

import (
	"context"
	"log/slog"

	"github.com/thediveo/go-mntinfo"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/mounts"
	"github.com/thediveo/lxkns/ops/mountineer"
	"github.com/thediveo/lxkns/species"
)

// NamespacedMountPathMap maps mount namespaces identified by their namespace ID
// to their corresponding mount path maps.
type NamespacedMountPathMap map[species.NamespaceID]mounts.MountPathMap

// discoverFromMountinfo discovers the mount paths with their mount points in
// mount namespaces. API users must have opted in not only to this discovery
// step but must have also enabled discovery of mount namespaces. Otherwise,
// this step will be skipped.
func discoverFromMountinfo(_ species.NamespaceType, _ string, result *Result) {
	if !result.Options.DiscoverMounts {
		slog.Info("skipping discovery of namespaces", slog.String("src", "mountpaths,mountpoints"))
		return
	}
	if result.Options.NamespaceTypes&species.CLONE_NEWNS == 0 {
		slog.Warn("mount namespace discovery skipped, so skipping mount path and points discovery")
	}
	slog.Debug("discovering namespaces", slog.String("src", "mountpaths,mountpoints"))
	// For every discovered mount namespace read its mount points, then
	// determine the mount paths and the mount point visibility from this
	// information.
	result.Mounts = NamespacedMountPathMap{}
	debugEnabled := slog.Default().Enabled(context.Background(), slog.LevelDebug)
	mountpointtotal := 0
	for mntid, mountns := range result.Namespaces[model.MountNS] {
		mnteer, err := mountineer.NewWithMountNamespace(
			mountns,
			result.Namespaces[model.UserNS])
		if err != nil {
			slog.Error("could not discover mount points",
				slog.String("namespace", mountns.(model.NamespaceStringer).TypeIDString()),
				slog.String("err", err.Error()))
			continue
		}
		if debugEnabled {
			slog.Debug("reading mount point information from bind-mounted namespace",
				slog.String("namespace", mountns.(model.NamespaceStringer).TypeIDString()),
				slog.String("ref", refString(mountns, result)))
		}
		// Warp speed Mr Sulu, through the proc root wormhole!
		mountpoints := mntinfo.MountsOfPid(int(mnteer.PID()))
		mnteer.Close()
		if debugEnabled {
			slog.Debug("found further mounts inside mount namespace",
				slog.String("namespace", mountns.(model.NamespaceStringer).TypeIDString()),
				slog.Int("count", len(mountpoints)))
		}
		mountpointtotal += len(mountpoints)
		result.Mounts[mntid] = mounts.NewMountPathMap(mountpoints)
	}
	slog.Info("found mount namespaces",
		slog.Int("count", len(result.Mounts)),
		slog.Int("mountpoint_count", mountpointtotal))
}
