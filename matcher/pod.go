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
	"github.com/onsi/gomega/types"
	"github.com/thediveo/lxkns/decorator/kuhbernetes"
	"github.com/thediveo/lxkns/model"
)

// HavePodName succeeds if actual is a model.Group or *model.Group and the
// groups is a Kubernetes/k8s pod of the specified namespace/name.
func HavePodName(namespacedname string) types.GomegaMatcher {
	return WithTransform(func(actual interface{}) (model.Group, error) {
		switch group := actual.(type) {
		case model.Group:
			return group, nil
		case *model.Group:
			return *group, nil
		default:
			return model.Group{}, fmt.Errorf(
				"HaveNamedPod expects a model.Group or *model.Group, but got %T", actual)
		}
	}, And(
		HaveField("Name", namespacedname),
		HaveField("Type", Equal(kuhbernetes.PodGroupType))))
}

// BelongToNamedPod succeeds if actual is a model.Container or *model.Container
// and the container is grouped by a Kubernetes/k8s pod.
func BelongToNamedPod(namespacedname string) types.GomegaMatcher {
	return withContainer("BeInNamedPod",
		HaveField("Groups", ContainElement(HavePodName(namespacedname))))
}
