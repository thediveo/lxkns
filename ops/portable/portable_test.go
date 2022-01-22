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

package portable

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

var _ = Describe("portable reference", func() {

	It("does not Open() a zero portable reference", func() {
		portref := PortableReference{}
		Expect(portref.Open()).Error().To(HaveOccurred())
	})

	It("does not Open() an invalid path", func() {
		portref := PortableReference{Path: "/foobar"}
		Expect(portref.Open()).Error().To(HaveOccurred())
	})

	It("does not Open() something which isn't a namespace at all", func() {
		portref := PortableReference{
			Path: "/proc/self/ns",
			Type: species.CLONE_NEWNS,
		}
		Expect(portref.Open()).Error().To(MatchError(MatchRegexp(`NS_GET_TYPE`)))
	})

	It("does not Open() wrong type of namespace", func() {
		portref := PortableReference{
			Path: "/proc/self/ns/net",
			Type: species.CLONE_NEWNS,
		}
		Expect(portref.Open()).Error().To(MatchError(MatchRegexp(`type mismatch.+expected mnt.+got net`)))
	})

	It("does not Open() wrong namespace ID", func() {
		portref := PortableReference{
			Path: "/proc/self/ns/net",
			ID:   species.NamespaceIDfromInode(666),
		}
		Expect(portref.Open()).Error().To(MatchError(MatchRegexp(
			fmt.Sprintf(`ID mismatch.+expected :\[666\].+got :\[%d\]`, mynetnsid.Ino))))
	})

	It("does not Open() when reference process is gone", func() {
		portref := PortableReference{
			Path:      "/proc/self/ns/net",
			PID:       -1,
			Starttime: 0,
		}
		Expect(portref.Open()).Error().To(MatchError(`process PID -1 is gone`))
		proc := model.NewProcess(model.PIDType(os.Getpid()))
		portref = PortableReference{
			Path:      "/proc/self/ns/net",
			PID:       proc.PID,
			Starttime: 0,
		}
		Expect(portref.Open()).Error().To(MatchError(MatchRegexp(`process PID [[:digit:]]+ is gone`)))
	})

	It("Open()s with only the namespace ID given", func() {
		portref := PortableReference{ID: mynetnsid}
		ref, closer, err := portref.Open()
		Expect(err).To(Succeed())
		Expect(closer).NotTo(BeNil())
		closer()
		Expect(ref).NotTo(BeNil())
	})

	It("Open()s with path and process cross-check", func() {
		proc := model.NewProcess(model.PIDType(os.Getpid()))
		portref := PortableReference{
			Path:      "/proc/self/ns/net",
			Type:      species.CLONE_NEWNET,
			PID:       proc.PID,
			Starttime: proc.Starttime,
		}
		ref, closer, err := portref.Open()
		Expect(err).To(Succeed())
		Expect(closer).NotTo(BeNil())
		closer()
		Expect(ref).NotTo(BeNil())
	})

})
