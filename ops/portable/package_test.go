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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
)

func TestRelations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "lxkns/ops/portable package")
}

var mynetnsid species.NamespaceID

var _ = BeforeSuite(func() {
	var err error
	mynetnsid, err = ops.NamespacePath("/proc/self/ns/net").ID()
	Expect(err).To(Succeed())
})
