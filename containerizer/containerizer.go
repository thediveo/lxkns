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

// +build linux

package containerizer

import (
	"context"

	"github.com/thediveo/lxkns/model"
)

// Containerizer discovers containers and relates them to processes (and thus
// also to Linux-kernel namespaces). A Containerizer can optionally be passed to
// a namespace discovery via the discovery options; the containerizer then will
// be called in order to discover any "alive" containers.
type Containerizer interface {
	// Discover user-level "alive" containers.
	//
	// Please note that depending on the particular containerizer implementation
	// the context might be used or not used at all.
	Containers(ctx context.Context, procs model.ProcessTable, pidmap model.PIDMapper) []model.Container
	// Close and release all resources allocated by this Containerizer.
	Close()
}
