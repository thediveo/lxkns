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
	"errors"
	"io"
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

func assertInvNSError(err error) {
	var invnserr *InvalidNamespaceError
	ExpectWithOffset(1, errors.As(err, &invnserr)).To(BeTrue(), "not an 'invalid namespace' error")
}

func null() *os.File {
	fnull, err := os.Open("/dev/null")
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "broken /dev/null")
	return fnull
}

var _ = Describe("Namespaces", func() {

	It("descriptively fails to wrap an invalid file descriptor", func() {
		Expect(errof(typedNamespaceFileFromFd(NamespacePath("goobarr"), "", ^uint(0), 0, nil))).To(
			MatchError(MatchRegexp("invalid file descriptor -1")))
	})

	It("wraps namespace *os.Files", func() {
		Expect(NamespaceFile{}.String()).To(MatchRegexp("zero os.File"))

		f, err := NewNamespaceFile(os.Open("/foobar"))
		Expect(err).To(HaveOccurred())
		assertInvNSError(err)
		Expect(err).To(MatchError(MatchRegexp(`^lxkns: invalid namespace: open /foobar: no such file.+$`)))
		Expect(errors.Unwrap(err)).NotTo(BeNil())
		Expect(f).To(BeNil())

		fnull := null()
		defer fnull.Close()
		f, err = NewNamespaceFile(fnull, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Fd()).To(Equal(fnull.Fd()))

		_, err = namespaceFileFromFd(f, ^uint(0), nil)
		Expect(err).To(HaveOccurred())
		assertInvNSError(err)
		Expect(err).To(MatchError(MatchRegexp(`^.+lxkns: invalid namespace os.File.+$`)))
	})

	It("returns types of referenced namespaces", func() {
		Expect(errof(NamespacePath("/foobar").Type())).To(HaveOccurred())
		Expect(NewTypedNamespacePath("/foobar", species.CLONE_NEWNET).Type()).To(Equal(species.CLONE_NEWNET))
		Expect(NewTypedNamespacePath("/foobar", species.CLONE_NEWUSER).String()).To(
			MatchRegexp(`path /foobar, type user`))

		Expect(errof(NamespaceFd(-1).Type())).To(HaveOccurred())
		ref, err := NewTypedNamespaceFd(-1, species.CLONE_NEWNET)
		Expect(err).NotTo(HaveOccurred())
		Expect(ref.Type()).To(Equal(species.CLONE_NEWNET))

		f, err := NewNamespaceFile(os.Open("relations_test.go"))
		Expect(err).To(Succeed())
		defer f.Close()
		Expect(errof(f.Type())).To(HaveOccurred())

		Expect(NamespacePath("/proc/self/ns/user").Type()).To(Equal(species.CLONE_NEWUSER))

		f, err = NewNamespaceFile(os.Open("/proc/self/ns/ipc"))
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(NamespaceFd(f.Fd()).Type()).To(Equal(species.CLONE_NEWIPC))

		Expect(f.Type()).To(Equal(species.CLONE_NEWIPC))

		Expect(errof(NamespacePath("doc.go").Type()).Error()).To(MatchRegexp("invalid namespace operation NS_GET_TYPE.+inappropriate ioctl"))
	})

	It("returns identifiers of namespaces", func() {
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

	It("opens typed references", func() {
		ref, err := NewTypedNamespaceFd(0, species.CLONE_NEWNS)
		Expect(err).NotTo(HaveOccurred())
		oref, closer, err := ref.OpenTypedReference()
		Expect(err).NotTo(HaveOccurred())
		Expect(closer).NotTo(BeNil())
		Expect(oref.(*TypedNamespaceFd)).To(BeIdenticalTo(ref))

		fref := &NamespaceFile{*os.Stdout}
		_, _, err = fref.OpenTypedReference()
		Expect(err).To(MatchError(MatchRegexp("invalid namespace operation NS_GET_NSTYPE")))

		fref, err = NewNamespaceFile(os.Open("/proc/self/ns/net"))
		Expect(err).NotTo(HaveOccurred())
		oref, closer, err = fref.OpenTypedReference()
		Expect(err).NotTo(HaveOccurred())
		Expect(closer).NotTo(BeNil())
		Expect(oref).NotTo(BeNil())

		fnull := null()
		defer fnull.Close()
		tfref, err := NewTypedNamespaceFile(fnull, species.CLONE_NEWUSER)
		Expect(err).NotTo(HaveOccurred())
		oref, closer, err = tfref.OpenTypedReference()
		Expect(err).NotTo(HaveOccurred())
		Expect(closer).NotTo(BeNil())
		Expect(oref).NotTo(BeNil())

		fdref := NamespaceFd(0)
		_, _, err = fdref.OpenTypedReference()
		Expect(err).To(MatchError(MatchRegexp("invalid namespace operation")))

		fd, err := unix.Open("/proc/self/ns/net", unix.O_RDONLY, 0)
		Expect(err).NotTo(HaveOccurred())
		fdref = NamespaceFd(fd)
		oref, closer, err = fdref.OpenTypedReference()
		Expect(err).NotTo(HaveOccurred())
		Expect(closer).NotTo(BeNil())
		Expect(oref).NotTo(BeNil())
	})

	It("returns suitable file descriptors for referencing", func() {
		ref := NamespacePath("foobar")
		_, _, err := ref.NsFd()
		Expect(err).To(HaveOccurred())
		assertInvNSError(err)

		nsf, err := os.Open("/proc/self/ns/net")
		Expect(err).ToNot(HaveOccurred())
		defer nsf.Close()

		fd, closer, err := (&NamespaceFile{*nsf}).NsFd()
		Expect(err).ToNot(HaveOccurred())
		Expect(closer).ToNot(BeNil())
		defer closer()
		Expect(fd).To(Equal(int(nsf.Fd())))

		fd, closer, err = NamespaceFd(nsf.Fd()).NsFd()
		Expect(err).ToNot(HaveOccurred())
		Expect(closer).ToNot(BeNil())
		defer closer()
		Expect(fd).To(Equal(int(nsf.Fd())))

		fd, closer, err = NamespacePath("/proc/self/ns/net").NsFd()
		Expect(err).ToNot(HaveOccurred())
		Expect(closer).ToNot(BeNil())
		defer closer()
		Expect(fd).ToNot(BeZero())
	})

	It("returns owning user namespaces", func() {
		Expect(errof(NamespacePath("/foo").User())).To(HaveOccurred())
		Expect(errof(NamespacePath("/").User())).To(HaveOccurred())
		Expect(errof(NamespaceFd(0).User())).To(HaveOccurred())

		usernsid, err := NamespacePath("/proc/self/ns/user").ID()
		Expect(err).ToNot(HaveOccurred())
		ownerfns, err := NamespacePath("/proc/self/ns/net").User()
		Expect(err).ToNot(HaveOccurred())
		defer ownerfns.(io.Closer).Close()
		Expect(ownerfns.ID()).To(Equal(usernsid))

		ownerf, err := NewNamespaceFile(os.Open("/proc/self/ns/net"))
		Expect(err).ToNot(HaveOccurred())
		defer ownerf.Close()
		userf, err := NamespaceFd(ownerf.Fd()).User()
		Expect(err).ToNot(HaveOccurred())
		defer userf.(io.Closer).Close()
		Expect(userf.ID()).To(Equal(usernsid))

		userf, err = ownerf.User()
		Expect(err).ToNot(HaveOccurred())
		defer userf.(io.Closer).Close()
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
		defer parentuserns.(io.Closer).Close()
		Expect(parentuserns.ID()).To(Equal(parentusernsid))
		pp, err := parentuserns.Parent()
		Expect(err).ToNot(HaveOccurred())
		defer pp.(io.Closer).Close()
		Expect(nstest.Err(pp.Parent())).To(HaveOccurred())

		leafuserf, err := os.Open(string(leafuserpath))
		Expect(err).ToNot(HaveOccurred())
		defer leafuserf.Close()
		parentuserns2, err := NamespaceFd(leafuserf.Fd()).Parent()
		Expect(err).ToNot(HaveOccurred())
		defer parentuserns2.(io.Closer).Close()
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

	It("tests helpers", func() {
		Expect(errof(NewTypedNamespaceFd(0, 0))).Should(HaveOccurred())
		ref, err := NewTypedNamespaceFd(42, species.CLONE_NEWNET)
		Expect(err).NotTo(HaveOccurred())
		Expect(ref.String()).To(MatchRegexp("fd 42.+type net"))

		Expect(errof(typedNamespaceFileFromFd(ref, "", 0, 0, errors.New("foobar")))).To(
			MatchError(MatchRegexp("invalid namespace fd")))
	})

})
