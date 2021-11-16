// Copyright 2021 Harald Albrecht.
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

package matcher

import (
	"fmt"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"github.com/thediveo/lxkns/model"
)

// HaveName succeeds if actual is a model.Container or *model.Container and the
// container matches the specified name or ID.
func HaveName(nameid string) types.GomegaMatcher {
	return withContainer(
		Or(
			HaveField("Name", Equal(nameid)),
			HaveField("ID", Equal(nameid))))
}

// HaveNameAndType succeeds if actual is a model.Container or *model.Container
// and the container matches the specified name or ID, as well as the specified
// type or flavor.
func HaveNameAndType(nameid string, typ string) types.GomegaMatcher {
	return withContainer(
		And(
			HaveName(nameid),
			Or(
				HaveField("Type", Equal(typ)),
				HaveField("Flavor", Equal(typ)))))
}

// HaveGroup succeeds if actual is a model.Container or *model.Container and the
// container is in a group with the specified name.
func HaveNamedGroup(name string) types.GomegaMatcher {
	return withContainer(
		HaveField("Groups", ContainElement(
			gstruct.PointTo(HaveField("Name", name)))))
}

// HaveNamedAndTypedGroup succeeds if actual is a model.Container or
// *model.Container and the container is in a group with the specified name and
// type or flavor.
func HaveNamedAndTypedGroup(name string, typ string) types.GomegaMatcher {
	return withContainer(
		HaveField("Groups", ContainElement(
			gstruct.PointTo(
				And(
					HaveField("Name", name),
					Or(
						HaveField("Type", typ),
						HaveField("Flavor", typ)))))))
}

// BePaused succeeds if actual is a model.Container or *model.Container and the
// container is paused.
func BePaused() types.GomegaMatcher {
	return withContainer(
		HaveField("Paused", BeTrue()))
}

// withContainer returns a matcher that transforms actual into a container value
// and then applied the specified matcher.
func withContainer(matcher types.GomegaMatcher) types.GomegaMatcher {
	return WithTransform(func(actual interface{}) (model.Container, error) {
		switch container := actual.(type) {
		case model.Container:
			return container, nil
		case *model.Container:
			return *container, nil
		default:
			return model.Container{}, fmt.Errorf(
				"HaveName expects a model.Container or *model.Container, but got %T", actual)
		}
	}, matcher)
}
