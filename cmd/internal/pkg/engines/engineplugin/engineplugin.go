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

package engineplugin

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/whalewatcher/watcher"
)

// NamedWatcher wraps a watcher and gives it a descriptive engine name.
type NamedWatcher struct {
	watcher.Watcher
	Name string // descriptive engine name
}

// NewWatchers is an exposed plugin function that returns a one or more (named)
// watcher(s) configured according to the CLI flags passed in, or nil if a
// particular engine should not be watched at all.
type NewWatchers func(cmd *cobra.Command) ([]*NamedWatcher, error)
