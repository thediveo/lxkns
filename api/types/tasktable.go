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

package types

import (
	"encoding/json"

	"github.com/thediveo/lxkns/model"
)

// TaskTable has no representation in the API but instead serves only for
// unmarshalling tasks.
type TaskTable map[model.PIDType]*model.Task

// Get always(!) returns a [model.Task] object for the specified task ID. When
// the task is already known, it is returned, else a new preliminary task object
// is created, registered, and returned. A preliminary task object has only its
// TID and the process it belongs to set.
//
// Please note that getting a previously unknown task object does not add it to
// the tasks of the process to which the task belongs. If needed, adding is the
// responsibility of the caller.
//
// A caller might not know the owning process yet, so it is allowed to specify
// it as nil. Yes, we're proudly nil-inclusive! If at a later stage the same
// task is requested again, but then the process is known, the reference from
// the task to its owning process is automatically updated.
func (t TaskTable) Get(proc *model.Process, tid model.PIDType) *model.Task {
	task, ok := t[tid]
	if !ok {
		task = &model.Task{
			TID:     tid,
			Process: proc,
		}
		t[tid] = task
	} else if task.Process == nil {
		task.Process = proc
	}
	return task
}

// Tasks is a list of (references to) Task objects.
type Tasks []*model.Task

// Task is the JSON representation of the information about a single task (of a
// process). This type is designed to be used under the hood of ProcessTable and
// not directly by 3rd (external) party users.
type Task model.Task

// MarshalJSON emits the textual JSON representation of a single task.
func (t *Task) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Namespaces *NamespacesSetReferences `json:"namespaces"`
		*model.Task
	}{
		Namespaces: (*NamespacesSetReferences)(&t.Namespaces),
		Task:       (*model.Task)(t),
	})
}

// UnmarshalJSON simply panics in order to clearly indicate that Task is
// not to be unmarshalled without a namespace dictionary to find existing
// namespaces in or add new ones just learnt to. Unfortunately, Golang's
// generic json (un)marshalling mechanism doesn't allow "contexts".
func (t *Task) UnmarshalJSON(data []byte) error {
	panic("cannot directly unmarshal github.com/thediveo/lxkns/api/types.Task")
}

// unmarshalJSON reads in the textual JSON representation of a single task. It
// uses the associated namespace dictionary to resolve existing references into
// namespace objects and also adds missing namespaces.
func (t *Task) unmarshalJSON(data []byte, allns *NamespacesDict) error {
	// While we unmarshal "most" of the task data using json's automated
	// mechanics, we need to deal with the namespaces a process is attached to
	// separately. Because we need context for the namespaces, we do it
	// manually and then extract only the parts we need here.
	aux := struct {
		Namespaces json.RawMessage `json:"namespaces"`
		*model.Task
	}{
		Task: (*model.Task)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Unmarshal the Namespaces field that need further special treatment.
	if err := (*NamespacesSetReferences)(&t.Namespaces).unmarshalJSON(aux.Namespaces, allns); err != nil {
		return err
	}
	return nil
}
