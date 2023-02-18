// Copyright 2023 Harald Albrecht.
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

package broken

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/cmd/internal/pkg/engines/engineplugin"
)

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Group[engineplugin.NewWatchers]().Register(
		NewWatchers, plugger.WithPlugin("broken-engine"))
}

// NewWatchers always returns an error, because this plugin in broken. 666 ...
// or so.
func NewWatchers(_ *cobra.Command) ([]*engineplugin.NamedWatcher, error) {
	return nil, errors.New("broken engine plugin")
}
