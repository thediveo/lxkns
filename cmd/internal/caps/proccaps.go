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

package caps

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/thediveo/caps"
	"github.com/thediveo/lxkns/model"
)

// ProcessCapabilities returns the names for the set of effective capabilities
// of the process specified by pid. The capability names are lower case and in
// lexicographic order. If there is an error determining the effective
// capabilities, then a nil slice is returned.
func ProcessCapabilities(pid model.PIDType) []string {
	return processEffectiveCaps(pid)
}

// processEffectiveCaps returns the names of the effective capabilities of
// process pid. The capability names are lower case and in lexicographic order.
// If there is an error determining the effective capabilities, then a nil slice
// is returned.
func processEffectiveCaps(pid model.PIDType) []string {
	f, err := os.Open("/proc/" + strconv.FormatUint(uint64(pid), 10) + "/status")
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()
	effcaps := statusEffectiveCaps(f)
	if effcaps == nil {
		return nil
	}
	capnames := effcaps.SortedNames()
	for idx := range capnames {
		capnames[idx] = strings.ToLower(capnames[idx])
	}
	return capnames
}

// statusEffectiveCaps reads the given process status file, extracts the
// effective capabilities from it and returns them as a caps.CapabilitiesSet. In
// case there is any error reading the effective capabilities, nil is returned
// instead.
func statusEffectiveCaps(f *os.File) caps.CapabilitiesSet {
	scanner := bufio.NewScanner(f)
	// Scan through the process status information until we arrive at the
	// sought-after "CapEff:" field. That's the only field interesting to us.
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "CapEff:\t") {
			capeff := strings.Split(line, "\t")[1]
			caps, err := caps.CapabilitiesFromHex(capeff)
			if err != nil {
				return nil
			}
			return caps
		}
	}
	return nil
}
