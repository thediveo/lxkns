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
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/thediveo/lxkns/model"
)

// ProcessCapabilities returns the set of effective capabilities of the process
// specified by pid.
func ProcessCapabilities(pid model.PIDType) []string {
	return capsToNames(processEffectiveCaps(pid))
}

// processEffectiveCaps returns the effective capabilities of process pid as
// []byte, with the least significant byte, erm octet, first.
func processEffectiveCaps(pid model.PIDType) (b []byte) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	return statusEffectiveCaps(f)
}

// capsToNames returns a slice of capabilities names (identifiers) in lower case
// for the specified capabilities byte string. The first capability is
// represented by bit 0 (=0x01) in the first byte.
func capsToNames(caps []byte) (capnames []string) {
	capbit := 0
	for _, b := range caps {
		for bit := byte(1); bit != 0; bit <<= 1 {
			if b&bit != 0 {
				if capname, ok := CapNames[capbit]; ok {
					capnames = append(capnames, capname)
				} else {
					capnames = append(capnames, fmt.Sprintf("cap_%d", capbit))
				}
			}
			capbit++
		}
	}
	sort.Strings(capnames)
	return capnames
}

// statusEffectiveCaps reads the given process status file, extracts the
// effective capabilities from it and returns them as a slice of bytes, with the
// LSB (least significant byte) first. That is, the first byte in the returned
// slice covers capabilities #0-#7, the second byte then covers caps #8-#15, et
// cetera. Capability #0 is bit 0 (=0x01), and so on. In case there is any error
// reading the effective capabilities, statusEffectiveCaps simply returns an
// empty slice.
func statusEffectiveCaps(f *os.File) (b []byte) {
	scanner := bufio.NewScanner(f)
	// Scan through the process status information until we arrive at the
	// sought-after "CapEff:" field. That's the only field interesting to us.
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "CapEff:\t") {
			capeff := strings.Split(line, "\t")[1]
			// Convert the hex string representation into a byte string and then
			// reverse it, so that the first byte holds caps #0-#7, and so on.
			var err error
			b, err = hex.DecodeString(capeff)
			if err != nil {
				// Ensure to reset the result, as DecodeString will return
				// whatever it could parse so far. And we don't want that.
				b = []byte{}
				return
			}
			for i := len(b)/2 - 1; i >= 0; i-- {
				opp := len(b) - 1 - i
				b[i], b[opp] = b[opp], b[i]
			}
			return
		}
	}
	return
}
