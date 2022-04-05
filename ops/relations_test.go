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

	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func assertInvNSError(err error) {
	var invnserr *InvalidNamespaceError
	Expect(errors.As(err, &invnserr)).WithOffset(1).To(BeTrue(), "not an 'invalid namespace' error")
}

func null() *os.File {
	fnull, err := os.Open("/dev/null")
	Expect(err).WithOffset(1).To(Succeed(), "broken /dev/null")
	return fnull
}

var _ = Describe("Namespaces", func() {

	It("descriptively fails to wrap an invalid file descriptor", func() {
		Expect(typedNamespaceFileFromFd(NamespacePath("goobarr"), "", ^uint(0), 0, nil)).Error().To(
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
		Expect(err).To(Succeed())
		Expect(f.Fd()).To(Equal(fnull.Fd()))

		_, err = namespaceFileFromFd(f, ^uint(0), nil)
		Expect(err).To(HaveOccurred())
		assertInvNSError(err)
		Expect(err).To(MatchError(MatchRegexp(`^.+lxkns: invalid namespace os.File.+$`)))
	})

	It("returns types of referenced namespaces", func() {
		Expect(NamespacePath("/foobar").Type()).Error().To(HaveOccurred())
		Expect(NewTypedNamespacePath("/foobar", species.CLONE_NEWNET).Type()).To(Equal(species.CLONE_NEWNET))
		Expect(NewTypedNamespacePath("/foobar", species.CLONE_NEWUSER).String()).To(
			MatchRegexp(`path /foobar, type user`))

		Expect(NamespaceFd(-1).Type()).Error().To(HaveOccurred())
		ref, err := NewTypedNamespaceFd(-1, species.CLONE_NEWNET)
		Expect(err).To(Succeed())
		Expect(ref.Type()).To(Equal(species.CLONE_NEWNET))

		f, err := NewNamespaceFile(os.Open("relations_test.go"))
		Expect(err).To(Succeed())
		defer f.Close()
		Expect(f.Type()).Error().To(HaveOccurred())

		Expect(NamespacePath("/proc/self/ns/user").Type()).To(Equal(species.CLONE_NEWUSER))

		f, err = NewNamespaceFile(os.Open("/proc/self/ns/ipc"))
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(NamespaceFd(f.Fd()).Type()).To(Equal(species.CLONE_NEWIPC))

		Expect(f.Type()).To(Equal(species.CLONE_NEWIPC))

		Expect(NamespacePath("doc.go").Type()).Error().To(MatchError(MatchRegexp("invalid namespace operation NS_GET_TYPE.+inappropriate ioctl")))
	})

	It("returns identifiers of namespaces", func() {
		Expect(NamespacePath("/foobar").ID()).Error().To(HaveOccurred())
		Expect(NamespaceFd(-1).ID()).Error().To(HaveOccurred())
		nsf, err := NewNamespaceFile(os.Open("/proc/self/ns/net"))
		Expect(err).ToNot(HaveOccurred())
		nsf.Close() // sic! make Fstat fail, that's why it is called "F"stat...
		Expect(nsf.ID()).Error().To(HaveOccurred())

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
		Expect(err).To(Succeed())
		oref, closer, err := ref.OpenTypedReference()
		Expect(err).To(Succeed())
		Expect(closer).NotTo(BeNil())
		Expect(closer).NotTo(Panic())
		Expect(oref.(*TypedNamespaceFd)).To(BeIdenticalTo(ref))

		fref := &NamespaceFile{*os.Stdout}
		Expect(fref.OpenTypedReference()).Error().To(MatchError(MatchRegexp("invalid namespace operation NS_GET_NSTYPE")))

		fref, err = NewNamespaceFile(os.Open("/proc/self/ns/net"))
		Expect(err).To(Succeed())
		oref, closer, err = fref.OpenTypedReference()
		Expect(err).To(Succeed())
		Expect(closer).NotTo(BeNil())
		Expect(closer).NotTo(Panic())
		Expect(oref).NotTo(BeNil())

		fnull := null()
		defer fnull.Close()
		tfref, err := NewTypedNamespaceFile(fnull, species.CLONE_NEWUSER)
		Expect(err).To(Succeed())
		oref, closer, err = tfref.OpenTypedReference()
		Expect(err).To(Succeed())
		Expect(closer).NotTo(BeNil())
		Expect(closer).NotTo(Panic())
		Expect(oref).NotTo(BeNil())

		fdref := NamespaceFd(0)
		Expect(fdref.OpenTypedReference()).Error().To(MatchError(MatchRegexp("invalid namespace operation")))

		fd, err := unix.Open("/proc/self/ns/net", unix.O_RDONLY, 0)
		Expect(err).To(Succeed())
		fdref = NamespaceFd(fd)
		oref, closer, err = fdref.OpenTypedReference()
		Expect(err).To(Succeed())
		Expect(closer).NotTo(BeNil())
		Expect(closer).NotTo(Panic())
		Expect(oref).NotTo(BeNil())

		Expect(NewTypedNamespacePath("foobar", 0).OpenTypedReference()).Error().To(
			MatchError(MatchRegexp("invalid namespace path")))
		Expect(NewTypedNamespacePath("doc.go", 0).OpenTypedReference()).Error().To(
			MatchError(MatchRegexp("invalid namespace path.+invalid namespace operation")))
		pref, closer, err := NewTypedNamespacePath("/proc/self/ns/net", 0).OpenTypedReference()
		Expect(err).To(Succeed())
		Expect(closer).NotTo(BeNil())
		Expect(closer).NotTo(Panic())
		Expect(pref).NotTo(BeNil())
		closer()
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
		Expect(NamespacePath("/foo").User()).Error().To(HaveOccurred())
		Expect(NamespacePath("/").User()).Error().To(HaveOccurred())
		Expect(NamespaceFd(0).User()).Error().To(HaveOccurred())

		usernsid, err := NamespacePath("/proc/self/ns/user").ID()
		Expect(err).ToNot(HaveOccurred())
		ownerfns, err := NamespacePath("/proc/self/ns/net").User()
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = ownerfns.(io.Closer).Close() }()
		Expect(ownerfns.ID()).To(Equal(usernsid))

		ownerf, err := NewNamespaceFile(os.Open("/proc/self/ns/net"))
		Expect(err).ToNot(HaveOccurred())
		defer ownerf.Close()
		userf, err := NamespaceFd(ownerf.Fd()).User()
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = userf.(io.Closer).Close() }()
		Expect(userf.ID()).To(Equal(usernsid))

		userf, err = ownerf.User()
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = userf.(io.Closer).Close() }()
		Expect(userf.ID()).To(Equal(usernsid))
	})

	It("returns the parent of a user namespace", func() {
		Expect(NamespacePath("/proc/self/ns/foobar").Parent()).Error().To(HaveOccurred())

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
		cmd.Decode(&parentuserpath)
		parentusernsid := nstest.CmdDecodeNSId(cmd)
		cmd.Decode(&leafuserpath)
		_ = nstest.CmdDecodeNSId(cmd)

		parentuserns, err := leafuserpath.Parent()
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = parentuserns.(io.Closer).Close() }()
		Expect(parentuserns.ID()).To(Equal(parentusernsid))
		pp, err := parentuserns.Parent()
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = pp.(io.Closer).Close() }()
		Expect(pp.Parent()).Error().To(HaveOccurred())

		Expect(NewTypedNamespacePath("foobar", 0).Parent()).Error().To(
			MatchError(MatchRegexp("invalid namespace path")))

		parentuserns, err = NewTypedNamespacePath(string(leafuserpath), species.CLONE_NEWUSER).Parent()
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = parentuserns.(io.Closer).Close() }()
		Expect(parentuserns.ID()).To(Equal(parentusernsid))

		f, err := os.Open(string(leafuserpath))
		Expect(err).To(Succeed())
		tleafuserns, err := NewTypedNamespaceFile(f, species.CLONE_NEWUSER)
		Expect(err).To(Succeed())
		parentuserns, err = tleafuserns.Parent()
		Expect(err).To(Succeed())
		defer func() { _ = parentuserns.(io.Closer).Close() }()
		Expect(parentuserns.ID()).To(Equal(parentusernsid))

		leafuserf, err := os.Open(string(leafuserpath))
		Expect(err).ToNot(HaveOccurred())
		defer leafuserf.Close()
		parentuserns2, err := NamespaceFd(leafuserf.Fd()).Parent()
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = parentuserns2.(io.Closer).Close() }()
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
		Expect(err).To(Succeed())
		defer f.Close()
		Expect(NamespaceFile{*f}.OwnerUID()).To(Equal(os.Getuid()))

		Expect(NamespaceFd(f.Fd()).OwnerUID()).To(Equal(os.Getuid()))

		Expect(NamespaceFd(0).OwnerUID()).Error().To(HaveOccurred())
		Expect(NamespacePath("/foo").OwnerUID()).Error().To(HaveOccurred())
	})

	It("tests helpers", func() {
		Expect(NewTypedNamespaceFd(0, 0)).Error().To(HaveOccurred())
		ref, err := NewTypedNamespaceFd(42, species.CLONE_NEWNET)
		Expect(err).To(Succeed())
		Expect(ref.String()).To(MatchRegexp("fd 42.+type net"))

		Expect(typedNamespaceFileFromFd(ref, "", 0, 0, errors.New("foobar"))).Error().To(
			MatchError(MatchRegexp("invalid namespace fd")))
	})

})
