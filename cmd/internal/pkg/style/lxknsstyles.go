// Defines the rendering styles used by lxkns tools when rendering specific
// elements, such as different namespace styles, process names, PIDs, et
// cetera.

// Copyright 2020 Harald Albrecht.
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

package style

// The set of styles for styling types of Linux-kernel namespaces differently,
// as well as some more elements, such as process names, user names, et
// cetera. The styles are meant to be directly referenced (used) by other
// packages importing our cmd/internal/style package
var (
	MntStyle    Style // styles mnt: namespaces
	CgroupStyle Style // styles cgroup: namespaces
	UTSStyle    Style // styles uts: namespaces
	IPCStyle    Style // styles ipc: namespaces
	UserStyle   Style // styles user: namespaces
	PIDStyle    Style // styles pid: namespaces
	NetStyle    Style // styles net: namespaces
	TimeStyle   Style // styles time: namespaces

	UserNoCapsStyle   Style // user: namespaces without capabilities
	UserEffCapsStyle  Style // user: namespaces with effective capabilities
	UserFullCapsStyle Style // user: namespaces with full capabilities

	OwnerStyle        Style // styles owner username and UID
	ProcessStyle      Style // styles process names
	ControlGroupStyle Style // control group names/references
	UnknownStyle      Style // styles undetermined elements, such as unknown PIDs.
)

// Styles maps style configuration top-level element names to their
// corresponding Style objects for storing and using specific style information.
var Styles = map[string]*Style{
	"mnt":    &MntStyle,
	"cgroup": &CgroupStyle,
	"uts":    &UTSStyle,
	"ipc":    &IPCStyle,
	"user":   &UserStyle,
	"pid":    &PIDStyle,
	"net":    &NetStyle,
	"time":   &TimeStyle,

	"user-nocaps":   &UserNoCapsStyle,
	"user-effcaps":  &UserEffCapsStyle,
	"user-fullcaps": &UserFullCapsStyle,

	"owner":        &OwnerStyle,
	"process":      &ProcessStyle,
	"controlgroup": &ControlGroupStyle,
	"unknown":      &UnknownStyle,
}
