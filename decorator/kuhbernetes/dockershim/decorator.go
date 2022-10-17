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

package dockershim

// TODO: sandbox container indication
// TODO: UID labelling

import (
	"regexp"
	"strings"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/decorator/kuhbernetes"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
)

// Register this Decorator plugin.
func init() {
	plugger.Group[decorator.Decorate]().Register(
		Decorate, plugger.WithPlugin("dockershim"))
}

const sandboxName = "POD"

const dnsLabelMaxLen = 63

var dnsLabelRe = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
var uidRe = regexp.MustCompile(`^[-a-z0-9]+$`)

// validate tests a non-zero identifier string against a maximum length and
// regular expression, returning true if the identifier validates. Returns false
// otherwise.
func validate(s string, re *regexp.Regexp, maxlen int) bool {
	return (s != "") && (len(s) <= maxlen) && re.MatchString(s)
}

// Decorate decorates the discovered Docker containers with pod groups, where
// applicable.
func Decorate(engines []*model.ContainerEngine, labels map[string]string) {
	total := 0
	for _, engine := range engines {
		// Pods cannot span container engines ;)
		podgroups := map[string]*model.Group{}
		for _, container := range engine.Containers {
			// Is this a container that is managed by the k8s Docker shim?
			if !strings.HasPrefix(container.Name, "k8s_") {
				continue
			}
			fields := strings.Split(container.Name, "_")
			if len(fields) != 6 && len(fields) != 7 {
				continue
			}
			containerName := fields[1]
			podName := fields[2]
			podNamespace := fields[3]
			podUid := fields[4]
			if (containerName != sandboxName && !validate(containerName, dnsLabelRe, dnsLabelMaxLen)) ||
				!validate(podName, dnsLabelRe, dnsLabelMaxLen) ||
				!validate(podNamespace, dnsLabelRe, dnsLabelMaxLen) ||
				!validate(podUid, uidRe, int((^uint(0))>>1)) {
				continue
			}
			// Add the container name as set by user as additional label.
			container.Labels[kuhbernetes.PodContainerNameLabel] = containerName
			// Create a new pod group, if it doesn't exist yet. Add the
			// container to its pod group.
			namespacedpodname := podNamespace + "/" + podName
			podgroup, ok := podgroups[namespacedpodname]
			if !ok {
				podgroup = &model.Group{
					Name:   namespacedpodname,
					Type:   kuhbernetes.PodGroupType,
					Flavor: kuhbernetes.PodGroupType,
				}
				podgroups[namespacedpodname] = podgroup
				total++
			}
			podgroup.AddContainer(container)
			// Is this container a sandbox? Then mark (label) it.
			if containerName == sandboxName {
				container.Labels[kuhbernetes.PodSandboxLabel] = ""
			}
			// Also label the UID we found in the dockershim container name.
			container.Labels[kuhbernetes.PodUidLabel] = podUid
		}
	}
	if total > 0 {
		log.Infof("discovered %d dockershim pods", total)
	}
}
