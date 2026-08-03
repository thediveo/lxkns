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

package discover

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/thediveo/spacetest/mntns"
	"github.com/thediveo/spacetest/netns"
	"github.com/thediveo/spacetest/units"
	"golang.org/x/sys/unix"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("Discover from bind-mounts", func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
				ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("finds hidden hierarchical user namespaces", func(ctx context.Context) {
		const bmpath = "/tmp/lxkns-discover-discovery_bindmount_test"

		netnsfd := netns.NewTransient()
		mntnsfd, _ := mntns.NewTransient()
		mntns.Execute(mntnsfd, func() {
			mntns.MountTmpfs(64 * units.KiB)
			Expect(os.WriteFile(bmpath, nil, 0770)).To(Succeed())
			Expect(unix.Mount(
				fmt.Sprintf("/proc/self/fd/%d", netnsfd),
				bmpath,
				"",
				unix.MS_BIND|unix.MS_RDONLY,
				"",
			)).To(Succeed(),
				"cannot bind transient netns into /tmp inside transient mntns")
		})

		netnsid := species.NamespaceIDfromInode(netns.Ino(netnsfd))
		allns := Namespaces(WithStandardDiscovery())
		Expect(allns.Namespaces[model.NetNS]).To(HaveKey(netnsid))
	})

})
