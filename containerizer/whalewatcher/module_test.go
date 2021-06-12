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

package whalewatcher

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rxtst "github.com/thediveo/gons/reexec/testing"
)

func TestMain(m *testing.M) {
	// Ensure that the registered handler is run in the re-executed child.
	// This won't trigger the handler while we're in the parent. We're using
	// gons' very special coverage profiling support for re-execution.
	mm := &rxtst.M{M: m}
	os.Exit(mm.Run())
}

func TestContainerizerWhalewatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "lxkns/containerizer/whalewatcher package")
}
