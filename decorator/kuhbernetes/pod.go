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

package kuhbernetes

import "github.com/thediveo/whalewatcher/engineclient/cri"

// PodGroupType identifies container groups representing Kubernetes pods.
const PodGroupType = "io.kubernetes.pod"

// PodSandboxLabel marks a container as a sandbox (or "pause") container; it is
// present only on sandboxes and the label value is irrelevant.
const PodSandboxLabel = cri.PodSandboxLabel

// PodNameLabel specifies the pod name of a container.
const PodNameLabel = cri.PodNameLabel

// PodNamespaceLabel specifies the namespace of the pod a container is part of.
const PodNamespaceLabel = cri.PodNamespaceLabel

// PodContainerNameLabel specifies the name of a container inside a pod from the
// Kubernetes perspective.
const PodContainerNameLabel = cri.PodContainerNameLabel

// PodUidLabel specifies the UID of a pod (=group).
const PodUidLabel = cri.PodUidLabel
