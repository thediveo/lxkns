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

package lxkns

import (
	"encoding/json"
	"io"
	"os/exec"
	"syscall"
	"time"

	t "github.com/thediveo/lxkns/nstypes"
	r "github.com/thediveo/lxkns/relations"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Discover", func() {

	It("finds the namespaces lsns finds", func() {
		allns := Discover(FullDiscovery)
		for _, ns := range lsns() {
			nsidx := TypeIndex(t.NameToType(ns.Type))
			dns := allns.Namespaces[nsidx][ns.NS]
			Expect(dns).NotTo(BeZero())
			Expect(dns.LeaderPIDs()).To(ContainElement(PIDType(ns.PID)))
		}
	})

	It("finds hidden hierarchical user namespaces", func() {
		cmd := NewCmd(
			"unshare", "-Ur", "unshare", "-U",
			"bash", "-c",
			`readlink /proc/self/ns/user | sed -n -e 's/^.\+:\[\(.*\)\]/\1/p' && read`)
		defer cmd.Close()
		var usernsid t.NamespaceID
		cmd.Decode(&usernsid)
		allns := Discover(FullDiscovery)
		userns := allns.Namespaces[UserNS][usernsid].(Hierarchy)
		Expect(userns).NotTo(BeNil())
		ppusernsid, _ := r.ID("/proc/self/ns/user")
		Expect(userns.Parent().Parent().(Namespace).ID()).To(Equal(ppusernsid))
	})

})

type Cmd struct {
	cmd             *exec.Cmd
	stdinr, stdoutr *io.PipeReader
	stdinw, stdoutw *io.PipeWriter
	dec             *json.Decoder
}

func NewCmd(command string, args ...string) *Cmd {
	cmd := &Cmd{
		cmd: exec.Command(command, args...),
	}
	cmd.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	inReader, inWriter := io.Pipe()
	cmd.cmd.Stdin = inReader
	cmd.stdinr = inReader
	cmd.stdinw = inWriter
	outReader, outWriter := io.Pipe()
	cmd.cmd.Stdout = outWriter
	cmd.stdoutr = outReader
	cmd.stdoutw = outWriter
	cmd.dec = json.NewDecoder(cmd.stdoutr)
	cmd.cmd.Start()
	return cmd
}

// Close completes the command by sending it an ENTER input and then closing
// the input pipe to the command. Then close waits at most 2s for the command
// to finish its business. If the command passes the timeout, then it will be
// killed hard.
func (cmd *Cmd) Close() {
	cmd.Proceed()
	cmd.stdinw.Close()
	done := make(chan error)
	go func() { done <- cmd.cmd.Wait() }()
	select {
	case <-time.After(2 * time.Second):
		// And if thou'rt unwilling...
		cmd.cmd.Process.Kill()
	case <-done:
	}
	cmd.stdinr.Close()
	cmd.stdoutr.Close()
	cmd.stdoutw.Close()
}

// Proceed sends the command an ENTER input.
func (cmd *Cmd) Proceed() {
	cmd.stdinw.Write([]byte{0x0a})
}

// Decode reads JSON from the command's output and tries to decode it into the
// data element specified.
func (cmd *Cmd) Decode(v interface{}) {
	err := cmd.dec.Decode(v)
	if err != nil {
		panic(err)
	}
}

type lsnsentry struct {
	NS      t.NamespaceID `json:"ns"`
	Type    string        `json:"type"`
	NProcs  int           `json:"nprocs"`
	PID     PIDType       `json:"pid"`
	User    string        `json:"user"`
	Command string        `json:"command"`
}

type lsnsdata struct {
	Namespaces []lsnsentry `json:"namespaces"`
}

func lsns(opts ...string) []lsnsentry {
	out, err := exec.Command(
		"lsns",
		append([]string{"--json"}, opts...)...).Output()
	if err != nil {
		panic(err)
	}
	var res lsnsdata
	err = json.Unmarshal(out, &res)
	if err != nil {
		panic(err)
	}
	if len(res.Namespaces) == 0 {
		panic("error: no namespaces read")
	}
	return res.Namespaces
}
