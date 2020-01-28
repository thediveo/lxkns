// The "lsuns" CLI tool for listing user namespaces, and optionally the other
// namespaces they "own".

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

package main

import (
	"fmt"

	asciitree "github.com/TheDiveO/go-asciitree"
	"github.com/thediveo/lxkns"
	common "github.com/thediveo/lxkns/cmd/internal/pkg/shared"
)

func main() {
	allns := lxkns.Discover(lxkns.FullDiscovery)
	fmt.Println(
		asciitree.Render(
			allns.UserNSRoots, &UserNSVisitor{true}, common.NamespaceStyler))
}
