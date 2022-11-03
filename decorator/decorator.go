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

// Decorate the discovered containers with groups, group-related container
// labels, et cetera. Instead of passing in all containers as a flat list, the
// containers are implicitly specified through their responsible container
// engines in order to allow decorators to apply engine-specific optimizations.
// The decoration can optionally be controlled for supporting decorators through
// labels (key-value pairs) specified as part of the discovery process.
//
// This type doubles as an exposed plugin symbol type for use with [plugger/v3].
//
// [plugger/v3]: https://github.com/thediveo/go-plugger
type Decorate func(engines []*model.ContainerEngine, labels map[string]string)
