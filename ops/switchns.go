// Switching namespaces.

// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ops

import (
	"runtime"
	"syscall"

	"golang.org/x/sys/unix"
)

// ...
type Opener interface {
	Open() (fd int, close bool, err error)
}

// Go runs the specified function as a new Go routine and from a locked OS
// thread, while joined to the specified namespaces. When the specified function
// returns, its Go routine will also terminate and the underlying OS thread will
// be destroyed. This avoids subtle problems further down the road in case there
// were namespace switching issues which overwise would carry over into any code
// executed after invoking Go(). Go() returns nil if switching namespaces
// succeeded, else an error. Please note that Go() returns as soon as switching
// namespaces has finished. The specified function is then run in its own Go
// routine.
func Go(f func(), nsrefs ...Opener) error {
	started := make(chan error)
	go func() {
		// Lock, but never unlock the OS thread exclusively powering our Go
		// routine. This ensures that the Golang runtime will destroy the OS
		// thread and never attempts to reuse it.
		runtime.LockOSThread()
		// Switch our highly exclusive OS thread into the specified
		// namespaces...
		for _, nsref := range nsrefs {
			fd, close, err := nsref.Open()
			if err != nil {
				started <- err
				return // ex-terminate ;)
			}
			err = unix.Setns(fd, 0)
			if close {
				// Don't leak open file descriptors...
				syscall.Close(int(fd))
			}
			if err != nil {
				started <- err
				return
			}
		}
		// Our preparations are finally done, so let's call the desired function and
		// then call it a day.
		started <- nil
		f()
	}()
	// Wait for the goroutine to have finished switching namespaces and about to
	// invoke the specified function. We're lazy and are never closing the
	// channel, but it will get garbage collected anyway.
	return <-started
}

// Execute a function synchronously while switched into the specified
// namespaces, then returns the interface{} outcome of calling the specified
// function. If switching fails, Execute returns an error instead.
func Execute(f func() interface{}, nsrefs ...Opener) (interface{}, error) {
	result := make(chan interface{})
	if err := Go(func() {
		result <- f()
	}, nsrefs...); err != nil {
		return nil, err
	}
	return <-result, nil
}
