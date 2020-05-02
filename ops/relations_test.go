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

package ops

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
	"golang.org/x/sys/unix"
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

	It("wraps namespace *os.Files", func() {
		f, err := NewNamespaceFile(os.Open("/foobar"))
		Expect(err).To(HaveOccurred())
		Expect(f).To(BeNil())

		f, err = NewNamespaceFile(os.Stdout, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Fd()).To(Equal(os.Stdout.Fd()))

		_, err = namespaceFileFromFd(^uint(0), nil)
		Expect(err).To(HaveOccurred())
	})

	It("return their types", func() {
		Expect(errof(NamespacePath("/foobar").Type())).To(HaveOccurred())
		Expect(errof(NamespaceFd(-1).Type())).To(HaveOccurred())

		Expect(NamespacePath("/proc/self/ns/user").Type()).To(Equal(species.CLONE_NEWUSER))

		f, err := NewNamespaceFile(os.Open("/proc/self/ns/ipc"))
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(NamespaceFd(f.Fd()).Type()).To(Equal(species.CLONE_NEWIPC))

		Expect(f.Type()).To(Equal(species.CLONE_NEWIPC))
	})

	It("return their identifiers", func() {
		Expect(errof(NamespacePath("/foobar").ID())).To(HaveOccurred())
		Expect(errof(NamespaceFd(-1).ID())).To(HaveOccurred())
		nsf, err := NewNamespaceFile(os.Open("/proc/self/ns/net"))
		Expect(err).ToNot(HaveOccurred())
		nsf.Close() // sic! make Fstat fail, that's why it is called "F"stat...
		Expect(errof(nsf.ID())).To(HaveOccurred())

		var stat unix.Stat_t
		Expect(unix.Stat("/proc/self/ns/cgroup", &stat)).ToNot(HaveOccurred())
		nsid := species.NamespaceID{Dev: stat.Dev, Ino: stat.Ino}

		Expect(NamespacePath("/proc/self/ns/cgroup").ID()).To(Equal(nsid))

		f, err := NewNamespaceFile(os.Open("/proc/self/ns/cgroup"))
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(NamespaceFd(f.Fd()).ID()).To(Equal(nsid))
		Expect(f.ID()).To(Equal(nsid))
	})

	It("return a suitable file descriptor for referencing", func() {
		nsf, err := os.Open("/proc/self/ns/net")
		Expect(err).ToNot(HaveOccurred())
		defer nsf.Close()

		fd, close, err := (&NamespaceFile{*nsf}).Reference()
		Expect(err).ToNot(HaveOccurred())
		Expect(close).To(BeFalse())
		Expect(fd).To(Equal(int(nsf.Fd())))

		fd, close, err = NamespaceFd(nsf.Fd()).Reference()
		Expect(err).ToNot(HaveOccurred())
		Expect(close).To(BeFalse())
		Expect(fd).To(Equal(int(nsf.Fd())))

		fd, close, err = NamespacePath("/proc/self/ns/net").Reference()
		Expect(err).ToNot(HaveOccurred())
		Expect(close).To(BeTrue())
		defer unix.Close(fd)
		Expect(fd).ToNot(BeZero())
	})

	It("return their owning user namespace", func() {
		Expect(errof(NamespacePath("/foo").User())).To(HaveOccurred())
		Expect(errof(NamespacePath("/").User())).To(HaveOccurred())
		Expect(errof(NamespaceFd(0).User())).To(HaveOccurred())

		usernsid, err := NamespacePath("/proc/self/ns/user").ID()
		Expect(err).ToNot(HaveOccurred())
		ownerfns, err := NamespacePath("/proc/self/ns/net").User()
		Expect(err).ToNot(HaveOccurred())
		defer ownerfns.Close()
		Expect(ownerfns.ID()).To(Equal(usernsid))

		ownerf, err := NewNamespaceFile(os.Open("/proc/self/ns/net"))
		Expect(err).ToNot(HaveOccurred())
		defer ownerf.Close()
		userf, err := NamespaceFd(ownerf.Fd()).User()
		Expect(err).ToNot(HaveOccurred())
		defer userf.Close()
		Expect(userf.ID()).To(Equal(usernsid))

		userf, err = ownerf.User()
		Expect(err).ToNot(HaveOccurred())
		defer userf.Close()
		Expect(userf.ID()).To(Equal(usernsid))
	})

	It("returns the parent of a user namespace", func() {
		Expect(errof(NamespacePath("/proc/self/ns/foobar").Parent())).To(HaveOccurred())

		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		// Creates a first user namespace: this will become the test's
		// "parent" user namespace. Then creates a second user namespace
		// inside the first user namespace. This then will become the leaf
		// user namespace.
		scripts.Script("newparent", `
unshare -Ur $parentuserns
`)
		scripts.Script("parentuserns", `
echo "\"/proc/$$/ns/user\""
process_namespaceid user
unshare -Uf $childuserns
`)
		scripts.Script("childuserns", `
echo "\"/proc/$$/ns/user\""
process_namespaceid user
read # wait for test to proceed()
`)
		cmd := scripts.Start("newparent")
		defer cmd.Close()

		var parentuserpath, leafuserpath NamespacePath
		var parentusernsid, leafusernsid species.NamespaceID
		cmd.Decode(&parentuserpath)
		cmd.Decode(&parentusernsid)
		cmd.Decode(&leafuserpath)
		cmd.Decode(&leafusernsid)

		parentuserns, err := leafuserpath.Parent()
		Expect(err).ToNot(HaveOccurred())
		defer parentuserns.Close()
		Expect(parentuserns.ID()).To(Equal(parentusernsid))
		pp, err := parentuserns.Parent()
		Expect(err).ToNot(HaveOccurred())
		defer pp.Close()
		Expect(nstest.Err(pp.Parent())).To(HaveOccurred())

		leafuserf, err := os.Open(string(leafuserpath))
		Expect(err).ToNot(HaveOccurred())
		defer leafuserf.Close()
		parentuserns2, err := NamespaceFd(leafuserf.Fd()).Parent()
		Expect(err).ToNot(HaveOccurred())
		defer parentuserns2.Close()
		Expect(parentuserns2.ID()).To(Equal(parentusernsid))
	})

	It("finds the owner UID", func() {
		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("newuserns", `
unshare -Ufr $userns
`)
		scripts.Script("userns", `
echo "\"/proc/$$/ns/user\""
read # wait for test to proceed()
`)
		cmd := scripts.Start("newuserns")
		defer cmd.Close()

		var userpath NamespacePath
		cmd.Decode(&userpath)

		uid, err := userpath.OwnerUID()
		Expect(err).ToNot(HaveOccurred())
		Expect(uid).To(Equal(os.Getuid()))

		f, err := os.Open(string(userpath))
		Expect(err).NotTo(HaveOccurred())
		defer f.Close()
		Expect(NamespaceFile{*f}.OwnerUID()).To(Equal(os.Getuid()))

		Expect(NamespaceFd(f.Fd()).OwnerUID()).To(Equal(os.Getuid()))

		Expect(nstest.Err(NamespaceFd(0).OwnerUID())).To(HaveOccurred())
		Expect(nstest.Err(NamespacePath("/foo").OwnerUID())).To(HaveOccurred())
	})

})
