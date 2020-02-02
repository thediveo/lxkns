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

package relations

import (
	"os"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/nstypes"
)

func errof(v ...interface{}) error {
	if len(v) != 2 {
		panic("expect exactly two return values")
	}
	if v[1] == nil {
		return nil
	}
	return v[1].(error)
}

var _ = Describe("Namespaces", func() {

	It("return their types", func() {
		Expect(errof(Type("/foobar"))).To(HaveOccurred())
		Expect(errof(Type("/"))).To(HaveOccurred())
		Expect(Type("/proc/self/ns/user")).To(Equal(nstypes.CLONE_NEWUSER))

		Expect(errof(Type(0))).To(HaveOccurred())
		f, err := os.Open("/proc/self/ns/ipc")
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(Type(f.Fd())).To(Equal(nstypes.CLONE_NEWIPC))

		Expect(Type(f)).To(Equal(nstypes.CLONE_NEWIPC))

		Expect(errof(Type(nil))).To(HaveOccurred())
	})

	It("return their identifiers", func() {
		Expect(errof(ID("/foobar"))).To(HaveOccurred())

		info, err := os.Stat("/proc/self/ns/cgroup")
		Expect(err).ToNot(HaveOccurred())
		stat, ok := info.Sys().(*syscall.Stat_t)
		Expect(ok).To(BeTrue())
		nsid := nstypes.NamespaceID(stat.Ino)

		Expect(ID("/proc/self/ns/cgroup")).To(Equal(nsid))

		Expect(errof(ID(-1))).To(HaveOccurred())
		f, err := os.Open("/proc/self/ns/cgroup")
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(ID(f.Fd())).To(Equal(nsid))

		Expect(ID(f)).To(Equal(nsid))

		Expect(errof(ID(nil))).To(HaveOccurred())
	})

	It("return their owning user namespace", func() {
		Expect(errof(User("/foo"))).To(HaveOccurred())
		Expect(errof(User("/"))).To(HaveOccurred())

		usernsid, err := ID("/proc/self/ns/user")
		Expect(err).ToNot(HaveOccurred())
		ownerf, err := User("/proc/self/ns/net")
		Expect(err).ToNot(HaveOccurred())
		defer ownerf.Close()
		Expect(ID(ownerf)).To(Equal(usernsid))

		ownerf, err = os.Open("/proc/self/ns/net")
		Expect(err).ToNot(HaveOccurred())
		defer ownerf.Close()
		userf, err := User(ownerf.Fd())
		Expect(err).ToNot(HaveOccurred())
		defer userf.Close()
		Expect(ID(userf)).To(Equal(usernsid))

		userf, err = User(ownerf)
		Expect(err).ToNot(HaveOccurred())
		defer userf.Close()
		Expect(ID(userf)).To(Equal(usernsid))

		Expect(errof(User(0))).To(HaveOccurred())
		Expect(errof(User(nil))).To(HaveOccurred())
	})

})
