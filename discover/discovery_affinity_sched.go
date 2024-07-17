// Copyright 2024 Harald Albrecht.
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

import "github.com/thediveo/lxkns/model"

// discoverAffinityScheduling discovers the CPU affinity lists and scheduler
// settings for the leader processes of all discovered namespaces.
func discoverAffinityScheduling(result *Result) {
	if !result.Options.DiscoverAffinityScheduling {
		return
	}
	for nstype := model.MountNS; nstype < model.NamespaceTypesCount; nstype++ {
		for _, ns := range result.Namespaces[nstype] {
			for _, leader := range ns.Leaders() {
				if leader.Affinity != nil {
					continue
				}
				_ = leader.RetrieveAffinityScheduling()
			}
		}
	}
}
