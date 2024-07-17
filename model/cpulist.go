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
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/unix"
)

// CPUList is a list of CPU [from...to] ranges. CPU numbers are starting from
// zero.
type CPUList [][2]uint

// CPUSet is a CPU bit string, such as used for CPU affinity masks. See also
// [sched_getaffinity(2)].
//
// [sched_getaffinity(2)]: https://man7.org/linux/man-pages/man2/sched_getaffinity.2.html
type CPUSet []uint64

// The dynamically determined size of CPUSets on this system (size in uint64
// words). This is usually smaller than the fixed-sized [unix.CPUSet] that Go's
// [unix.SchedGetaffinity] uses.
var setsize atomic.Uint64
var wordbytesize = uint64(unsafe.Sizeof(CPUSet{0}[0]))

func init() {
	setsize.Store(1)
}

// NewAffinityCPUList returns the affinity CPUList (list of CPU ranges) of the
// process with the passed PID. Otherwise, it returns an error. If pid is zero,
// then the affinity CPU list of the calling thread is returned (make sure to
// have the OS-level thread locked to the calling go routine in this case).
//
// Notes:
//   - we don't use [unix.SchedGetaffinity] as this is tied to the fixed size
//     [unix.CPUSet] type; instead, we dynamically figure out the size needed
//     and cache the size internally.
//   - retrieving the affinity CPU mask and then speed-running it to
//     generate the range list is roughly two orders of magnitude faster than
//     fetching “/proc/$PID/status” and looking for the “Cpus_allowed_list”,
//     because generating the broad status procfs file is expensive.
func NewAffinityCPUList(pid PIDType) (CPUList, error) {
	var set CPUSet

	setlenStart := setsize.Load()
	setlen := setlenStart
	for {
		set = make([]uint64, setlen)
		// see also:
		// https://man7.org/linux/man-pages/man2/sched_setaffinity.2.html; we
		// use RawSyscall here instead of Syscall as we know that
		// SYS_SCHED_GETAFFINITY does not block, following Go's stdlib
		// implementation.
		_, _, e := unix.RawSyscall(unix.SYS_SCHED_GETAFFINITY,
			uintptr(pid), uintptr(setlen*wordbytesize), uintptr(unsafe.Pointer(&set[0])))
		if e != 0 {
			if e == unix.EINVAL {
				setlen *= 2
				continue
			}
			return nil, e
		}
		// Set the new size; if this fails because another go routine already
		// upped the set size, retry until we either notice that we're smaller
		// than what was set as the new set size, or we succeed in setting the
		// size.
		for {
			if setsize.CompareAndSwap(setlenStart, setlen) {
				break
			}
			setlenStart = setsize.Load()
			if setlenStart > setlen {
				break
			}
		}
		break
	}
	return set.NewCPUList(), nil
}

// NewCPUList returns a list of CPU ranges for the given bitmap CPUSet.
//
// This is an optimized implementation that does not use any division and modulo
// operations; instead, it only uses increment and (single bit position) shift
// operations. Additionally, this implementation fast-forwards through all-0s
// and all-1s CPUSet words (uint64's).
func (s CPUSet) NewCPUList() CPUList {
	setlen := uint64(len(s))
	cpulist := CPUList{}
	cpuno := uint(0)
	cpuwordidx := uint64(0)
	cpuwordmask := uint64(1)

findNextCPUInWord:
	for {
		// If we're inside a cpu mask word, try to find the next set cpu bit, if
		// any, otherwise stop after we've fallen off the MSB end of the cpu
		// mask word.
		if cpuwordmask != 1 {
			for {
				if s[cpuwordidx]&cpuwordmask != 0 {
					break
				}
				cpuno++
				cpuwordmask <<= 1
				if cpuwordmask == 0 {
					// Oh no! We've fallen off the disc^Wcpu mask word.
					cpuwordidx++
					cpuwordmask = 1
					break
				}
			}
		}
		// Try to fast-forward through completely unset cpu mask words, where
		// possible.
		for cpuwordidx < setlen && s[cpuwordidx] == 0 {
			cpuno += 64
			cpuwordidx++
		}
		if cpuwordidx >= setlen {
			return cpulist
		}
		// We arrived at a non-zero cpu mask word, so let's now find the first
		// cpu in it.
		for {
			if s[cpuwordidx]&cpuwordmask != 0 {
				break
			}
			cpuno++
			cpuwordmask <<= 1
		}
		// Hooray! We've finally located a CPU in use. Move on to the next CPU,
		// handling a word boundary when necessary.
		cpufrom := cpuno
		cpuno++
		cpuwordmask <<= 1
		if cpuwordmask == 0 {
			// Oh no! We've again fallen off the disc^Wcpu mask word.
			cpuwordidx++
			cpuwordmask = 1
		}
		// Now locate the next unset cpu within the currently inspected cpu mask
		// word, until we find one or have exhausted our search within the
		// current cpu mask word.
		if cpuwordmask != 1 {
			for {
				if s[cpuwordidx]&cpuwordmask == 0 {
					cpulist = append(cpulist, [2]uint{cpufrom, cpuno - 1})
					continue findNextCPUInWord
				}
				cpuno++
				cpuwordmask <<= 1
				if cpuwordmask == 0 {
					cpuwordidx++
					cpuwordmask = 1
					break
				}
			}
		}
		// Try to fast-forward through completely set cpu mask words, where
		// applicable.
		for cpuwordidx < setlen && s[cpuwordidx] == ^uint64(0) {
			cpuno += 64
			cpuwordidx++
		}
		// Are we completely done? If so, add the final CPU span and then call
		// it a day.
		if cpuwordidx >= setlen {
			cpulist = append(cpulist, [2]uint{cpufrom, cpuno - 1})
			return cpulist
		}
		// We arrived at a non-all-1s cpu mask word, so let's now find the first
		// cpu in it that is unset. Add the CPU span, and then rinse and repeat
		// from the beginning: find the next set CPU or fall off the disc.
		for {
			if s[cpuwordidx]&cpuwordmask == 0 {
				cpulist = append(cpulist, [2]uint{cpufrom, cpuno - 1})
				break
			}
			cpuno++
			cpuwordmask <<= 1
		}
	}
}
