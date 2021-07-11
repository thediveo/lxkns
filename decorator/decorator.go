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

package decorator

import "github.com/thediveo/lxkns/model"

// PluginGroup specifies the name of the plugger group for decorator plugins.
const PluginGroup = "lxkns/plugingroup/decorator"

// Decorate processes the discovered containers from container engines,
// decorating containers with groups, et cetera.
type Decorate func(engines []*model.ContainerEngine)
