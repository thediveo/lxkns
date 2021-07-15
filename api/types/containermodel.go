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

package types

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/thediveo/lxkns/model"
)

// ContainerModel wraps the discovery information model part consisting of
// containers, their container engines and groups for (un)marshalling from/to
// JSON.
type ContainerModel struct {
	Containers       ContainerMap
	ContainerEngines EngineMap
	Groups           GroupMap
}

// NewContainerModel returns a new ContainerModel for (un)marshalling,
// optionally preparing it from a list of discovered containers (with managing
// container engines and groups).
func NewContainerModel(containers []*model.Container) *ContainerModel {
	cm := &ContainerModel{}
	cm.Containers = NewContainerMap(cm, containers)
	cm.ContainerEngines = NewEngineMap(cm, containers)
	cm.Groups = NewGroupMap(cm, containers)
	return cm
}

// ContainerMap wraps a set of discovered model.Containers for JSON
// (un)marshalling.
type ContainerMap struct {
	Containers map[uint]*model.Container // map ref IDs to containers.
	cm         *ContainerModel
}

// NewContainerMap returns a ContainerMap optionally initialized from a set of
// model.Containers.
func NewContainerMap(cm *ContainerModel, containers []*model.Container) ContainerMap {
	m := ContainerMap{
		Containers: map[uint]*model.Container{},
		cm:         cm,
	}
	for _, container := range containers {
		m.Containers[uint(container.PID)] = container
	}
	return m
}

// ContainerSlice returns the containers stored in the ContainerMap.
func (m ContainerMap) ContainerSlice() []*model.Container {
	containers := make([]*model.Container, 0, len(m.Containers))
	for _, container := range m.Containers {
		containers = append(containers, container)
	}
	return containers
}

// ContainerByRefID returns the Container object identified by the specified
// (ref) ID. If the object isn't yet known, a new zero'd object is returned.
func (m ContainerMap) ContainerByRefID(refid uint) *model.Container {
	container, ok := m.Containers[refid]
	if !ok {
		container = &model.Container{}
		m.Containers[refid] = container
	}
	return container
}

// ContainerMarshal is a model.Container with additional fields for
// (un)marshalling the engine and group references, as we cannot directly
// serialize plain pointers in an information model with lots of cycles.
type ContainerMarshal struct {
	Engine uint   `json:"engine"` // engine ref IDs.
	Groups []uint `json:"groups"` // group ref IDs.
	*model.Container
}

// MarshalJSON emits a set of containers in JSON textual format, representing
// the original object pointers to container engines and groups with ID
// references (in form of numbers).
func (m *ContainerMap) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	b.WriteRune('{')
	first := true
	for refid, container := range m.Containers {
		if first {
			first = false
		} else {
			b.WriteRune(',')
		}
		b.WriteRune('"')
		b.WriteString(strconv.FormatUint(uint64(refid), 10))
		b.WriteString(`":`)
		gids := make([]uint, len(container.Groups))
		for idx, group := range container.Groups {
			gids[idx] = m.cm.Groups.GroupRefID(group)
		}
		cntrjson, err := json.Marshal(&ContainerMarshal{
			Engine:    m.cm.ContainerEngines.EngineRefID(container.Engine),
			Groups:    gids,
			Container: container,
		})
		if err != nil {
			return nil, err
		}
		b.Write(cntrjson)
	}
	b.WriteRune('}')
	return b.Bytes(), nil
}

// UnmarshalJSON converts the JSON textual format back into a set of containers,
// including resolving container engine and group IDs into object references.
// Depending on a particular order of unmarshalling containers, engines, and
// groups, preliminary zero'd container engine and group objects are created, to
// be filled later as unmarshalling progresses.
func (m *ContainerMap) UnmarshalJSON(data []byte) error {
	aux := map[uint]json.RawMessage{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	for refid, rawc := range aux {
		container := m.ContainerByRefID(refid)
		caux := ContainerMarshal{
			Container: container,
		}
		if err := json.Unmarshal(rawc, &caux); err != nil {
			return err
		}
		container.Engine = m.cm.ContainerEngines.EngineByRefID(caux.Engine)
		groups := make([]*model.Group, len(caux.Groups))
		for idx, gid := range caux.Groups {
			groups[idx] = m.cm.Groups.GroupByRefID(gid)
		}
		container.Groups = groups
	}
	return nil
}

// EngineMap wraps a set of discovered model.ContainerEngines for JSON
// (un)marshalling.
type EngineMap struct {
	enginesByRefID map[uint]*model.ContainerEngine // map ref IDs to engines.
	engineRefIDs   map[*model.ContainerEngine]uint // associate (ref) IDs with the engines.
	cm             *ContainerModel
}

// NewEngineMap creates a new map for ContainerEngines, optionally building
// using a discovered list of containers (with their ContainerEngines).
func NewEngineMap(cm *ContainerModel, containers []*model.Container) EngineMap {
	m := EngineMap{
		enginesByRefID: map[uint]*model.ContainerEngine{},
		engineRefIDs:   map[*model.ContainerEngine]uint{},
		cm:             cm,
	}
	// If containers were discovered, then associate (ref) IDs with the engines
	// managing the containers.
	eid := uint(0)
	for _, container := range containers {
		if _, ok := m.engineRefIDs[container.Engine]; !ok {
			eid++
			m.engineRefIDs[container.Engine] = eid // associate a new ID with the engine
			m.enginesByRefID[eid] = container.Engine
		}
	}
	return m
}

// EngineByRefID returns the ContainerEngine associated with a (ref) ID,
// creating a new zero ContainerEngine if necessary.
func (m EngineMap) EngineByRefID(refid uint) *model.ContainerEngine {
	engine, ok := m.enginesByRefID[refid]
	if !ok {
		engine = &model.ContainerEngine{}
		m.enginesByRefID[refid] = engine
	}
	return engine
}

// EngineRefID returns the (ref) ID associated with a particular
// ContainerEngine.
func (m EngineMap) EngineRefID(engine *model.ContainerEngine) uint {
	return m.engineRefIDs[engine]
}

// EngineMarshal is a model.ContainerEngine with additional fields for
// (un)marshalling the container references, as we cannot directly serialize
// plain pointers in an information model with lots of cycles.
type EngineMarshal struct {
	Containers []uint `json:"containers"` // container ref IDs.
	*model.ContainerEngine
}

// MarshalJSON emits a set of container engines in JSON textual format,
// representing the original object pointers to containers with ID references
// (in form of numbers).
func (m *EngineMap) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	b.WriteRune('{')
	first := true
	for refid, engine := range m.enginesByRefID {
		if first {
			first = false
		} else {
			b.WriteRune(',')
		}
		b.WriteRune('"')
		b.WriteString(strconv.FormatUint(uint64(refid), 10))
		b.WriteString(`":`)
		cids := make([]uint, len(engine.Containers))
		for idx, container := range engine.Containers {
			cids[idx] = uint(container.PID)
		}
		engjson, err := json.Marshal(&EngineMarshal{
			Containers:      cids,
			ContainerEngine: (*model.ContainerEngine)(engine),
		})
		if err != nil {
			return nil, err
		}
		b.Write(engjson)
	}
	b.WriteRune('}')
	return b.Bytes(), nil
}

// UnmarshalJSON converts the JSON textual format back into a set of container
// engines, including resolving container IDs into object references. Depending
// on a particular order of unmarshalling containers, engines, and groups,
// preliminary zero'd container objects are created, to be filled later as
// unmarshalling progresses.
func (m *EngineMap) UnmarshalJSON(data []byte) error {
	aux := map[uint]json.RawMessage{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	for refid, rawe := range aux {
		engine := m.EngineByRefID(refid)
		eaux := EngineMarshal{
			ContainerEngine: engine,
		}
		if err := json.Unmarshal(rawe, &eaux); err != nil {
			return err
		}
		containers := make([]*model.Container, len(eaux.Containers))
		for idx, cid := range eaux.Containers {
			containers[idx] = m.cm.Containers.ContainerByRefID(cid)
		}
		engine.Containers = containers
	}
	return nil
}

// GroupMap wraps a set of discovered model.Groups for JSON (un)marshalling.
type GroupMap struct {
	groupsByRefID map[uint]*model.Group // map ref IDs to groups.
	groupRefIDs   map[*model.Group]uint // associate (ref) IDs with the groups.
	cm            *ContainerModel
}

// NewEngineMap creates a new map for ContainerEngines, optionally building
// using a discovered list of containers (with their ContainerEngines).
func NewGroupMap(cm *ContainerModel, containers []*model.Container) GroupMap {
	m := GroupMap{
		groupsByRefID: map[uint]*model.Group{},
		groupRefIDs:   map[*model.Group]uint{},
		cm:            cm,
	}
	// If containers were discovered, then associate (ref) IDs with the groups
	// grouping these containers.
	gid := uint(0)
	for _, container := range containers {
		for _, group := range container.Groups {
			if _, ok := m.groupRefIDs[group]; !ok {
				gid++
				m.groupRefIDs[group] = gid // associate a new ID with the group
				m.groupsByRefID[gid] = group
			}
		}
	}
	return m
}

// GroupByRefID returns the ContainerEngine associated with a (ref) ID, creating
// a new zero Group if necessary.
func (m GroupMap) GroupByRefID(refid uint) *model.Group {
	group, ok := m.groupsByRefID[refid]
	if !ok {
		group = &model.Group{}
		m.groupsByRefID[refid] = group
	}
	return group
}

// GroupRefID returns the (ref) ID associated with a particular Group.
func (m GroupMap) GroupRefID(group *model.Group) uint {
	return m.groupRefIDs[group]
}

// GroupMarshal is a model.Group with additional fields for (un)marshalling the
// container references, as we cannot directly serialize plain pointers in an
// information model with lots of cycles.
type GroupMarshal struct {
	Containers []uint `json:"containers"` // container ref IDs.
	*model.Group
}

// MarshalJSON emits a set of groups in JSON textual format, representing the
// original object pointers to containers with ID references (in form of
// numbers).
func (m *GroupMap) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	b.WriteRune('{')
	first := true
	for refid, group := range m.groupsByRefID {
		if first {
			first = false
		} else {
			b.WriteRune(',')
		}
		b.WriteRune('"')
		b.WriteString(strconv.FormatUint(uint64(refid), 10))
		b.WriteString(`":`)
		cids := make([]uint, len(group.Containers))
		for idx, container := range group.Containers {
			cids[idx] = uint(container.PID)
		}
		engjson, err := json.Marshal(&GroupMarshal{
			Containers: cids,
			Group:      (*model.Group)(group),
		})
		if err != nil {
			return nil, err
		}
		b.Write(engjson)
	}
	b.WriteRune('}')
	return b.Bytes(), nil
}

// UnmarshalJSON converts the JSON textual format back into a set of groups,
// including resolving container IDs into object references. Depending on a
// particular order of unmarshalling containers, engines, and groups,
// preliminary zero'd container objects are created, to be filled later as
// unmarshalling progresses.
func (m *GroupMap) UnmarshalJSON(data []byte) error {
	aux := map[uint]json.RawMessage{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	for refid, rawg := range aux {
		group := m.GroupByRefID(refid)
		gaux := GroupMarshal{
			Group: group,
		}
		if err := json.Unmarshal(rawg, &gaux); err != nil {
			return err
		}
		containers := make([]*model.Container, len(gaux.Containers))
		for idx, cid := range gaux.Containers {
			containers[idx] = m.cm.Containers.ContainerByRefID(cid)
		}
		group.Containers = containers
	}
	return nil
}
