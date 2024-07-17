// Copyright 2024 Harald Albrecht.
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

package model

import (
	"os"
	"strconv"
	"strings"
)

// SortProcessByPID sorts processes by increasing PID numbers (no interval
// arithmetics though).
func SortProcessByPID(a, b *Process) int {
	return int(a.PID) - int(b.PID)
}

// SortProcessByAgeThenPIDDistance sorts processes first by their “age”
// (starttime) and then by their PIDs, taking PID number wrap-arounds into
// consideration.
//
// As PIDs are monotonously increasing, wrapping around at “N” (which defaults
// to 1<<22 on Linux 64 bit systems), we consider a PID “B” to be after PID “A”
// if the “positive” distance from “A” to “B” (in increasing PIDs, distance
// taken modulo N) is at most N/2.
//
// For a nice write-up see also [The ryg blog: Intervals in modular arithmetic].
//
// [The ryg blog: Intervals in modular arithmetic]: https://fgiesen.wordpress.com/2015/09/24/intervals-in-modular-arithmetic/
func SortProcessByAgeThenPIDDistance(a, b *Process) int {
	switch {
	case a.Starttime < b.Starttime:
		return -1
	case a.Starttime > b.Starttime:
		return 1
	}
	pidA := uint64(a.PID)
	pidB := uint64(b.PID)
	switch dist := (pidB - pidA) & pidMaxMask; {
	case dist == 0:
		return 0
	case dist <= pidMaxDist:
		return -1
	default:
		return 1
	}
}

var pidMaxMask uint64 // N-1
var pidMaxDist uint64 // N/2

// pidWrapping reads the PID interval “N” set for this system (which must be to
// the power of two) and then returns N-1 and N/2, falling back to the specified
// default N in case the system configuration cannot be read.
func pidWrapping(defaultMax uint64) (mask, maxdist uint64) {
	mask = defaultMax - 1
	maxdist = defaultMax >> 1
	// https://www.man7.org/linux/man-pages/man5/proc_sys_kernel.5.html
	pidmaxb, err := os.ReadFile("/proc/sys/kernel/pid_max")
	if err != nil {
		return
	}
	pidmax, err := strconv.ParseUint(strings.TrimSuffix(string(pidmaxb), "\n"), 10, 32)
	if err != nil {
		return
	}
	return pidmax - 1, pidmax >> 1
}

func init() {
	pidMaxMask, pidMaxDist = pidWrapping((uint64(1) << 22))
}
