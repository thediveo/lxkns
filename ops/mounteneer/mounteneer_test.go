// Copyright 2021 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mounteneer

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("mounteneer", func() {

	It("opens a mount namespace path in initial context", func() {

		/*
			// Don't want to run a full discovery, just prime the user namespace map
			// with the needed entry.
			usernsmap := model.NamespaceMap{}
			usernsid, err := ops.NamespacePath(fmt.Sprintf("/proc/%d/ns/user", pid)).ID()
			Expect(err).NotTo(HaveOccurred())
			usernsmap[usernsid] = namespaces.New(species.CLONE_NEWUSER, usernsid, fmt.Sprintf("/proc/%d/ns/user", pid))
		*/

		// FIXME: needs a correct test setup.
		m, err := New([]string{"/run/snapd/ns/chromium.mnt"}, nil)
		Expect(err).NotTo(HaveOccurred())
		defer m.Close()

		// It must have created a sandbox/pause process.
		Expect(m.sandbox).NotTo(BeNil())
		// And the sandbox must not have terminated even if waiting a few
		// moments.
		Consistently(func() *os.ProcessState {
			return m.sandbox.ProcessState
		}, "1s", "250ms").Should(BeNil())
		// the contentsroot must be set to the sandbox process.
		Expect(m.contentsRoot).To(Equal(
			fmt.Sprintf("/proc/%d/root", m.sandbox.Process.Pid)))

		// Content path resolution must be correct.
		path, err := m.Resolve("/writable")
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal(
			fmt.Sprintf("/proc/%d/root/writable", m.sandbox.Process.Pid)))

		root, err := m.Resolve("/")
		Expect(err).NotTo(HaveOccurred())
		files, err := ioutil.ReadDir(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).NotTo(BeEmpty())

		// Correctly stops the sandbox process -- no sandbox leaks, please.
		m.Close()
		Eventually(func() *os.ProcessState {
			return m.sandbox.ProcessState
		}, "1s", "250ms").ShouldNot(BeNil())
	})

})
