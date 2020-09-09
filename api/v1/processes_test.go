package v1

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/species"
)

var _ = Describe("JSON", func() {

	It("TypedNamespaceSet un/marshalling", func() {
		// To JSON infinity...
		dummyns := lxkns.NewNamespace(species.CLONE_NEWNET, species.NamespaceID{Dev: 666, Ino: 123}, "")
		dummyuns := lxkns.NewNamespace(species.CLONE_NEWUSER, species.NamespaceID{Dev: 666, Ino: 456}, "")
		nsset := TypedNamespacesSet{
			nil, dummyns, dummyns, dummyns, dummyuns, dummyns, nil, dummyns,
		}
		j, err := json.Marshal(nsset)
		Expect(err).NotTo(HaveOccurred())
		Expect(j).To(MatchJSON(`{"user":456,"uts":123,"cgroup":123,"ipc":123,"pid":123,"time":123}`))

		// ...and back again!
		var tnsset TypedNamespacesSet
		Expect(json.Unmarshal(j, &tnsset)).NotTo(HaveOccurred())
		// FIXME: check only types & IDs
		// Expect(tnsset).To(Equal(nsset))
	})

	It("ProcessTable un/marshalling", func() {
		proc1 := &lxkns.Process{
			PID:       1,
			PPID:      0,
			Cmdline:   []string{"/sbin/domination", "--world"},
			Name:      "(init)",
			Starttime: 123,
			// TODO: Namespaces
		}
		proc2 := &lxkns.Process{
			PID:       666,
			PPID:      proc1.PID,
			Cmdline:   []string{"/sbin/fool"},
			Name:      "fool",
			Starttime: 666666,
			// TODO: Namespaces
		}
		procs := ProcessTable{
			proc1.PID: proc1,
			proc2.PID: proc2,
		}
		j, err := json.Marshal(procs)
		Expect(err).NotTo(HaveOccurred())
		Expect(j).To(MatchJSON(`""`))
	})

})
