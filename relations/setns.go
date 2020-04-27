package relations

import (
	"runtime"
	"syscall"

	"github.com/thediveo/lxkns/nstypes"
)

// ...
type Opener interface {
	Open() (fd uintptr, close bool, err error)
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
			err = Setns(fd, 0)
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

// Implements missing syscall.Setns.
func Setns(fd uintptr, nst nstypes.NamespaceType) error {
	_, _, e1 := syscall.Syscall(SYS_SETNS, fd, uintptr(nst), 0)
	if e1 != 0 {
		return e1
	}
	return nil
}

// CPU architecture-specific Linux syscall (trap) numbers taken from:
// https://github.com/vishvananda/netns/blob/master/netns_linux.go
var SYS_SETNS = map[string]uintptr{
	"386":      346,
	"amd64":    308,
	"arm64":    268,
	"arm":      375,
	"mips":     4344,
	"mipsle":   4344,
	"mips64le": 4344,
	"ppc64":    350,
	"ppc64le":  350,
	"riscv64":  268,
	"s390x":    339,
}[runtime.GOARCH]
