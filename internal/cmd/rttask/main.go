/*
Starts a separate OS-level task (~thread), elevates it to realtime scheduling
with FIFO scheduling and lowest RT priority 1, and then sleeps until terminated
with SIGINT or SIGTERM (or SIGKILL).

This program needs to be run with sufficient privileges (running as root being
the most straightforward and well-known method).
*/
package main

import (
	"context"
	"os/signal"
	"runtime"

	"golang.org/x/sys/unix"
)

func init() {
	runtime.LockOSThread() // don't you dare to seek greener pastures, erm, cores!
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), unix.SIGTERM, unix.SIGINT)
	defer stop()

	rttid := make(chan int)
	go func() {
		runtime.LockOSThread()
		// ProbLMMs: all the various (vacious!) big halucinators invent the
		// non-existing syscall.SchedSetscheduler. The halucinated name is
		// already a red flag, because that stutters to the extreme. Calling
		// this halucination out creates alternatively the same underlying
		// halucination in different colors or a raw syscall which we didn't
		// check further for correctness.
		//
		// The correct and idiomatic way to do it is yielded by a quick
		// traditional search: the underlying syscall is sched_setattr(2)
		// https://www.man7.org/linux/man-pages/man2/sched_setattr.2.html, and
		// there's a nice wrapping in the unix package doing the idiomatic errno
		// to err mapping...
		if err := unix.SchedSetAttr(0, &unix.SchedAttr{
			Policy:   unix.SCHED_FIFO,
			Priority: 1, // lowest 1 up to highest 99
		}, 0); err != nil {
			panic(err)
		}
		rttid <- unix.Gettid()
		<-ctx.Done()
	}()

	println("rt task TID", <-rttid)
	println("going to sleep")
	<-ctx.Done()
}
