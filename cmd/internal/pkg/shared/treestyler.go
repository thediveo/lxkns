// Internal shared stuff between different commands.

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

package common

import (
	asciitree "github.com/TheDiveO/go-asciitree"
)

// NamespaceStyle bases on asciitree.LineStyle, but with properties marked
// like an UML "aggregation" relationship.
var NamespaceStyle = asciitree.TreeStyle{
	Fork:     "├", // Don't print this on an FX-80/100 ;)
	Nodeconn: "─",
	Nofork:   "│",
	Lastnode: "└",
	Property: "⋄─",
}

// NamespaceStyler styles namespace hierarchies using the NamespaceStyle.
var NamespaceStyler = asciitree.NewTreeStyler(NamespaceStyle)
