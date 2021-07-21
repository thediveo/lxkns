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

package containerdtest

import (
	"os"

	"github.com/containerd/containerd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const testref = "docker.io/library/busybox:latest"

var testargs = []string{"/bin/sleep", "120s"}

var _ = Describe("creates and destroys test containers", func() {

	var pool *Pool

	BeforeEach(func() {
		if os.Geteuid() != 0 {
			Skip("needs root")
		}
		var err error
		pool, err = NewPool("/proc/1/root/run/containerd/containerd.sock", "containerd-test")
		Expect(err).NotTo(HaveOccurred())
	})

	It("doesn't fail when purging non-existing container", func() {
		pool.PurgeID("sluggish_sleepy")
		pool.PurgeID("sluggish_sleepy")
	})

	It("creates running container and destroys it", func() {
		// Make sure the image isn't cached yet, to be on the safe side.
		_ = pool.Client.ImageService().Delete(pool.context(), testref)
		// Make sure the slate is clean...
		pool.PurgeID("sluggish_sleepy")

		c, err := pool.Run("sluggish_sleepy", testref, true, testargs)
		Expect(err).NotTo(HaveOccurred())
		Expect(c).NotTo(BeNil())
		defer pool.Purge(c)

		Expect(c.Status()).To(Equal(containerd.Running))
	})

	It("creates paused container", func() {
		pool.PurgeID("sluggish_sleepy")
		c, err := pool.Run("sluggish_sleepy", testref, false, testargs)
		Expect(err).NotTo(HaveOccurred())
		defer pool.Purge(c)

		Expect(c.Status()).To(Equal(containerd.Paused))
	})

})
