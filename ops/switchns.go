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

	"golang.org/x/sys/unix"
)

// Referrer returns an open file descriptor to the namespace indicated in a
// namespace reference type, such as NamespacePath, suitable for switching
// namespaces using setns(2).
type Referrer interface {
	// Open returns a file descriptor referencing the namespace indicated in a
	// namespace reference implementing the Opener interface. If the returned
	// close is true, then the caller must close the file descriptor after it
	// doesn't need it anymore. If false, the caller must not close the file
	// descriptor. In case the Opener is unable to return a file descriptor to
	// the referenced namespace, err is non-nil.
	//
	// The caller must make sure that the namespace reference object does not
	// get garbage collected before the file descriptor is used, if in doubt,
	// use runtime.KeepAlive(nsref), see also:
	// https://golang.org/pkg/runtime/#KeepAlive.
	Reference() (fd int, close bool, err error)
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
func Go(f func(), nsrefs ...Referrer) error {
	started := make(chan error)
	go func() {
		// Lock, but never unlock the OS thread exclusively powering our Go
		// routine. This ensures that the Golang runtime will destroy the OS
		// thread and never attempts to reuse it.
		runtime.LockOSThread()
		// Switch our highly exclusive OS thread into the specified
		// namespaces...
		for _, nsref := range nsrefs {
			// Important: since nsref.Reference() returns a file descriptor
			// which potentially is derived from an open os.File, the latter
			// must not get garbage collected while we attempt to use the file
			// descriptor, as otherwise the os.File's finalizer will have closed
			// the fd prematurely. Luckily (hopefully not!) the (varargs) slice
			// won't be collectible until the iteration terminates, keeping its
			// slice elements and thus its os.Files (if any) alive. In
			// consequence, we don't need an explicit runtime.KeepAlive(...)
			// here.
			fd, close, err := nsref.Reference()
			if err != nil {
				started <- err
				return // ex-terminate ;)
			}
			err = unix.Setns(fd, 0)
			if close {
				// Don't leak open file descriptors...
				unix.Close(int(fd))
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
func Execute(f func() interface{}, nsrefs ...Referrer) (interface{}, error) {
	result := make(chan interface{})
	if err := Go(func() {
		result <- f()
	}, nsrefs...); err != nil {
		return nil, err
	}
	return <-result, nil
}
