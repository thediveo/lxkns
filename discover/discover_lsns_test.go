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

package discover

import (
	"encoding/json"
	"os/exec"
	"strconv"

	"github.com/thediveo/lxkns/model"
)

// lsnsentry represents the JSON information for individual namespaces spit
// out by "lsns --json", specifically in what might be the "v2" JSON schema.
// The older "v1" JSON schema serializes all properties as strings instead,
// including the Number/integer-typed elements.
type lsnsentry struct {
	NS      uint64        `json:"ns"`
	Type    string        `json:"type"`
	NProcs  int           `json:"nprocs"`
	PID     model.PIDType `json:"pid"`
	User    string        `json:"user"`
	Command string        `json:"command"`
}

// lsnsdata represents the JSON top-level element spit out by "lsns --json".
type lsnsdata struct {
	Namespaces []lsnsentry `json:"namespaces"`
}

// UnmarshalJSON does custom JSON unmarshalling in order to hide the
// differences between the lsns v1 and v2 JSON schemas.
func (e *lsnsentry) UnmarshalJSON(b []byte) (err error) {
	var fields map[string]*json.RawMessage
	if err = json.Unmarshal(b, &fields); err != nil {
		return
	}
	var i uint64
	if err = touint64(fields["ns"], &e.NS); err != nil {
		return
	}
	if err = tostr(fields["type"], &e.Type); err != nil {
		return
	}
	if err = touint64(fields["nprocs"], &i); err != nil {
		return
	}
	e.NProcs = int(i)
	if err = touint64(fields["pid"], &i); err != nil {
		return
	}
	e.PID = model.PIDType(i)
	if err = tostr(fields["user"], &e.User); err != nil {
		return
	}
	err = tostr(fields["command"], &e.Command)
	return
}

// tostr is a simple JSON unmarshalling convenience function which is just for
// symmetry with its slightly more complex toint() sibling.
func tostr(r *json.RawMessage, v *string) (err error) {
	err = json.Unmarshal(*r, v)
	return
}

// touint64 unmarshalles either a JSON number or string into a Golang int.
func touint64(r *json.RawMessage, v *uint64) (err error) {
	var s string
	if err = json.Unmarshal(*r, &s); err == nil {
		*v, err = strconv.ParseUint(s, 10, 64)
		return err
	}
	err = json.Unmarshal(*r, v)
	return
}

// lsns runs the "lsns" CLI command with the "--json" argument and then
// collects its JSON output and returns the data as a slice of []lsnsentry.
func lsns(opts ...string) []lsnsentry {
	out, err := exec.Command(
		"lsns",
		append([]string{"--json"}, opts...)...).Output()
	if err != nil {
		panic(err)
	}
	var res lsnsdata
	if err = json.Unmarshal(out, &res); err != nil {
		panic(err.Error())
	}
	if len(res.Namespaces) == 0 {
		panic("error: no namespaces read")
	}
	return res.Namespaces
}
