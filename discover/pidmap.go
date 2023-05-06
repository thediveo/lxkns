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

//go:build linux

package discover

import (
	ipidmap "github.com/thediveo/lxkns/internal/pidmap"
	"github.com/thediveo/lxkns/model"
)

// NewPIDMap returns a new PID map ([model.PIDMapper]) based on the specified
// discovery results and further information gathered from the /proc filesystem.
func NewPIDMap(result *Result) model.PIDMapper {
	return ipidmap.NewPIDMap(result.Processes)
}
