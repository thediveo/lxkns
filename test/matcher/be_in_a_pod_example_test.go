// Copyright 2022 Harald Albrecht.
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
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/decorator/kuhbernetes"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/engineclient/moby"
)

func ExampleBeInAPod() {
	var container = model.Container{
		ID:     "1234567890",
		Name:   "foo_bar",
		Type:   moby.Type,
		Flavor: moby.Type,
		Groups: []*model.Group{
			{
				Name:   "space/pod",
				Type:   kuhbernetes.PodGroupType,
				Flavor: kuhbernetes.PodGroupType,
			},
		},
	}
	container.Groups[0].Containers = []*model.Container{&container}
	pod := container.Groups[0]

	Expect(container).To(BeInAPod())
	Expect(container).To(BeInAPod(WithName(pod.Name)))
	Expect(container).NotTo(BeInAPod(WithName("Bielefeld")))
	//Output:
}
