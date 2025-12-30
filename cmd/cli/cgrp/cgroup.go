// Copyright 2020, 2026 Harald Albrecht.
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

package cgrp

import (
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	"github.com/thediveo/enumflag/v2"
	"github.com/thediveo/go-plugger/v3"
)

// Names of the CLI flags defined and used in this package.
const (
	CgroupFlagName = "cgroup"
)

// CgroupDisplayName returns a function configured based on CLI flags, where the
// returned function takes a control group name (path) and returns a name better
// suited for display. In particular, it optionally shortens 64 hex digit IDs as
// used by Docker for identifying containers to the Docker-typical 12 hex digit
// "digest".
func CgroupDisplayName(cmd *cobra.Command) func(string) string {
	cgroupNames := cmd.PersistentFlags().Lookup(CgroupFlagName).
		Value.(*enumflag.EnumFlagValue[ControlGroupNames]).GetValue()
	if cgroupNames == CgroupComplete {
		return func(s string) string { return s }
	}
	return func(s string) string {
		labels := strings.Split(s, "/")
		for idx, label := range labels {
			if len(label) != 64 || !ishex(label) {
				continue
			}
			labels[idx] = label[:12] + "…"
		}
		return strings.Join(labels, "/")
	}
}

// ishex checks if the given string solely consists of ASCII hex digits, and
// nothing else, then return true.
func ishex(hex string) bool {
	for _, char := range hex {
		if !unicode.In(char, unicode.ASCII_Hex_Digit) {
			return false
		}
	}
	return true
}

// ControlGroupNames defines the enumeration flag type for controlling
// optimizing control group names for display (or not).
type ControlGroupNames enumflag.Flag

const (
	// CgroupShortened enables optimizing the display of Docker container IDs.
	CgroupShortened ControlGroupNames = iota
	// CgroupComplete switches off any display optimization of control group
	// names.
	CgroupComplete
)

// ControlGroupNameModes specifies the mapping between the user-facing CLI flag
// values and the program-internal flag values.
var ControlGroupNameModes = map[ControlGroupNames][]string{
	CgroupShortened: {"short"},
	CgroupComplete:  {"full", "complete"},
}

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		SetupCLI, plugger.WithPlugin("lxkns/cgroup"))
}

// SetupCLI adds the flags for controlling control group name display.
func SetupCLI(cmd *cobra.Command) {
	cgroupNames := CgroupShortened
	cmd.PersistentFlags().Var(
		enumflag.New(&cgroupNames, "cgroup", ControlGroupNameModes, enumflag.EnumCaseInsensitive),
		CgroupFlagName,
		"control group name display; can be 'full' or 'short'")
}
