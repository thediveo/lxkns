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
	"fmt"
	"os/exec"
	"strconv"

	"github.com/thediveo/lxkns/nstest"
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
			discons := allns.Namespaces[nsidx][ns.NS]
			Expect(discons).NotTo(BeZero())
			// rats ... lsns seems to take the numerically lowest PID number
			// instead of the topmost PID in a namespace. This makes
			// Expect(dns.LeaderPIDs()).To(ContainElement(PIDType(ns.PID)))
			// give false negatives.
			p, ok := allns.Processes[ns.PID]
			Expect(ok).To(BeTrue(), "unknown PID %d", ns.PID)
			leaders := discons.LeaderPIDs()
			func() {
				pids := []PIDType{}
				for p != nil {
					pids = append(pids, p.PID)
					for _, lPID := range leaders {
						if lPID == p.PID {
							return
						}
					}
					p = p.Parent
				}
				Fail(fmt.Sprintf("PIDs %v not found in leaders %v", pids, leaders))
			}()
		}
	})

	It("finds hidden hierarchical user namespaces", func() {
		scripts := nstest.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("doubleunshare", `
unshare -Ur unshare -U $print_userns
`)
		scripts.Script("print-userns", `
process_namespaceid user
read # wait for test to proceed()
`)
		cmd := scripts.Start("doubleunshare")
		defer cmd.Close()
		var usernsid t.NamespaceID
		cmd.Decode(&usernsid)
		allns := Discover(FullDiscovery)
		userns := allns.Namespaces[UserNS][usernsid].(Hierarchy)
		Expect(userns).NotTo(BeNil())
		ppusernsid, _ := r.ID("/proc/self/ns/user")
		Expect(userns.Parent().Parent().(Namespace).ID()).To(Equal(ppusernsid))
	})

	It("rejects finding roots for plain namespaces", func() {
		opts := NoDiscovery
		opts.SkipProcs = false
		opts.NamespaceTypes = t.CLONE_NEWNET
		allns := Discover(opts)
		Expect(func() { rootNamespaces(allns.Namespaces[NetNS]) }).To(Panic())
	})

})

type lsnsentry struct {
	NS      t.NamespaceID `json:"ns"`
	Type    string        `json:"type"`
	NProcs  int           `json:"nprocs"`
	PID     PIDType       `json:"pid"`
	User    string        `json:"user"`
	Command string        `json:"command"`
}

func (e *lsnsentry) UnmarshalJSON(b []byte) (err error) {
	var fields map[string]*json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	var i int
	if err = toint(fields["ns"], &i); err != nil {
		return
	}
	e.NS = t.NamespaceID(i)
	if err = tostr(fields["type"], &e.Type); err != nil {
		return
	}
	if err = toint(fields["nprocs"], &e.NProcs); err != nil {
		return
	}
	if err = toint(fields["pid"], &i); err != nil {
		return
	}
	e.PID = PIDType(i)
	if err = tostr(fields["user"], &e.User); err != nil {
		return
	}
	err = tostr(fields["command"], &e.Command)
	return
}

func tostr(r *json.RawMessage, v *string) (err error) {
	err = json.Unmarshal(*r, v)
	return
}

func toint(r *json.RawMessage, v *int) (err error) {
	var s string
	if err = json.Unmarshal(*r, &s); err == nil {
		*v, err = strconv.Atoi(s)
		return
	}
	err = json.Unmarshal(*r, v)
	return
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
	if err = json.Unmarshal(out, &res); err != nil {
		panic(err.Error())
	}
	if len(res.Namespaces) == 0 {
		panic("error: no namespaces read")
	}
	return res.Namespaces
}
