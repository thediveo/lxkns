package gmodel

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

var allns *lxkns.DiscoveryResult
var userns1 model.Namespace
var proc1 *model.Process

var _ = BeforeSuite(func() {
	allns = lxkns.Discover(lxkns.FullDiscovery)
	proc1 = allns.Processes[model.PIDType(os.Getpid())]
	userns1 = proc1.Namespaces[model.UserNS]
})

var _ = Describe("Namespace", func() {

	It("matches same ID and Type", func() {
		ns1 := namespaces.New(species.CLONE_NEWNET, species.NamespaceIDfromInode(1), "/foobar")
		ns2 := namespaces.New(species.CLONE_NEWNET, species.NamespaceIDfromInode(1), "/foobar")
		Expect(ns1).To(EqualNamespace(ns2))
		Expect(ns1).NotTo(
			EqualNamespace(namespaces.New(species.CLONE_NEWUSER, species.NamespaceIDfromInode(1), "/foobar")))
		Expect(ns1).NotTo(
			EqualNamespace(namespaces.New(species.CLONE_NEWNET, species.NamespaceIDfromInode(666), "/foobar")))
		Expect(ns1).NotTo(
			EqualNamespace(namespaces.New(species.CLONE_NEWNET, species.NamespaceIDfromInode(1), "/foo/bar")))
	})

	It("matches PID lists", func() {
		Expect(sameLeaders([]model.PIDType{}, []model.PIDType{1})).To(BeFalse())
		Expect(sameLeaders([]model.PIDType{1}, []model.PIDType{2})).To(BeFalse())
		Expect(sameLeaders([]model.PIDType{1, 2}, []model.PIDType{1, 2})).To(BeTrue())
		Expect(sameLeaders([]model.PIDType{1, 2}, []model.PIDType{2, 1})).To(BeTrue())
	})

	It("matches namespaces", func() {
		Expect(nil).To(EqualNamespace(nil))
		_, err := EqualNamespace("bar").Match("foo")
		Expect(err).To(HaveOccurred())
		_, err = EqualNamespace("bar").Match(userns1)
		Expect(err).To(HaveOccurred())

		// ID, Type, and ref path must be checked
		Expect(userns1).NotTo(EqualNamespace(proc1.Namespaces[model.NetNS]))

		u2 := *(userns1.(*namespaces.UserNamespace))
		userns2 := model.Namespace(&u2)
		Expect(userns1).To(EqualNamespace(userns2))

		// leader PIDs must be checked
		u2 = *(userns1.(*namespaces.UserNamespace))
		userns2.(namespaces.NamespaceConfigurer).AddLeader(proc1)
		Expect(userns1).NotTo(EqualNamespace(userns2))

		// Guard ;) :p
		u2 = *(userns1.(*namespaces.UserNamespace))
		Expect(userns1).To(EqualNamespace(userns2))

		// parents must be checked
		u2 = *(userns1.(*namespaces.UserNamespace))
		dummyuserns := namespaces.New(species.CLONE_NEWUSER, species.NamespaceIDfromInode(1), "/roode")
		dummyuserns.(namespaces.HierarchyConfigurer).AddChild(model.Hierarchy(&u2))
		Expect(userns1).NotTo(EqualNamespace(userns2))

		// uid must be checked
		u2 = *(userns1.(*namespaces.UserNamespace))
		userns2.(namespaces.UserConfigurer).SetOwnerUID(123)
		Expect(userns1).NotTo(EqualNamespace(userns2))

		// children must be same...
		u0 := *(userns1.(*namespaces.UserNamespace))
		c1 := namespaces.New(species.CLONE_NEWUSER, species.NamespaceIDfromInode(667), "/667")
		c2 := namespaces.New(species.CLONE_NEWUSER, species.NamespaceIDfromInode(668), "/668")
		(namespaces.HierarchyConfigurer)(&u0).AddChild(c1.(model.Hierarchy))
		(namespaces.HierarchyConfigurer)(&u0).AddChild(c2.(model.Hierarchy))
		Expect(&u0).NotTo(EqualNamespace(userns1))

		u2 = u0
		Expect(&u2).To(EqualNamespace(&u0))

		u2 = *(userns1.(*namespaces.UserNamespace))
		(namespaces.HierarchyConfigurer)(&u0).AddChild(c2.(model.Hierarchy))
		(namespaces.HierarchyConfigurer)(&u0).AddChild(c1.(model.Hierarchy))
		Expect(&u2).To(EqualNamespace(&u0))
	})

})
