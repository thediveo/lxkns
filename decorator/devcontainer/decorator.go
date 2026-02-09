// Copyright 2026 Harald Albrecht.
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

package devcontainer

import (
	"encoding/json"
	"log/slog"
	"path/filepath"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/model"
)

const (
	CodespaceNameLabelName    = "lxkns.codespace.name"
	DevContainerNameLabelName = "lxkns.devcontainer.name"
)

const (
	DevcontainerConfigfileLabelName  = "devcontainer.config_file"
	DevcontainerLocalfolderLabelName = "devcontainer.local_folder"
	DevcontainerMetadataLabelName    = "devcontainer.metadata"
)

// Metadata for a devcontainer (including codespace) consists of a sequence of
// configuration blocks.
type Metadata []ConfigBlock

// ConfigBlock is our limited view on a configuration block, where we are solely
// interested in the presence of a “containerEnv” field that hopefully
// contains an object with “magic” environment variable fields.
type ConfigBlock struct {
	ContainerEnv CodespaceContainerEnv `json:"containerEnv"`
}

// CodespaceContainerEnv is our specialized view on a container environment
// definition with “magic” env variables that tell us more about a codespace.
type CodespaceContainerEnv struct {
	Codespaces     string `json:"CODESPACES"`     // "true"; yes, it's a string, not a because it's an env var.
	RepositoryName string `json:"RepositoryName"` // actually the codespace name, eh?!
}

// Register this Decorator plugin.
func init() {
	plugger.Group[decorator.Decorate]().Register(
		Decorate, plugger.WithPlugin("devcontainer"))
}

// Decorate all devcontainers, differentiating between “standalone”
// devcontainers and devcontainers in codespaces.
func Decorate(engines []*model.ContainerEngine, _ map[string]string) {
	total := 0
	for _, engine := range engines {
		for _, cntr := range engine.Containers {
			switch {
			case detectLocalDevContainer(cntr):
			case detectCodespace(cntr):
			}
		}
	}
	if total > 0 {
		slog.Info("discovered devcontainers", slog.Int("count", total))
	}
}

// detectLocalDevContainer adds a devcontainer name label in case of a "local"
// devcontainer that has a local folder path label attached and returns true,
// otherwise false.
func detectLocalDevContainer(cntr *model.Container) bool {
	localFolderPath := cntr.Labels[DevcontainerLocalfolderLabelName]
	if localFolderPath == "" {
		return false
	}
	name := filepath.Base(localFolderPath)
	if name == "" || name == "/" || name == "." {
		return false
	}
	cntr.Labels[DevContainerNameLabelName] = name
	return true
}

// detectCodespace adds a codespace name label if the specified container has to
// correct metadata label attached and returns true, otherwise false.
func detectCodespace(cntr *model.Container) bool {
	metadataLabel := cntr.Labels[DevcontainerMetadataLabelName]
	if metadataLabel == "" {
		return false
	}
	var metadata Metadata
	if err := json.Unmarshal([]byte(metadataLabel), &metadata); err != nil {
		return false
	}
	for _, block := range metadata {
		if block.ContainerEnv.Codespaces != "true" {
			continue
		}
		codespaceName := block.ContainerEnv.RepositoryName
		if codespaceName == "" {
			continue
		}
		cntr.Labels[CodespaceNameLabelName] = codespaceName
	}
	return true
}
