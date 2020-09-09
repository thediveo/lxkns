package v1

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/thediveo/lxkns"
)

// Processes is the JSON serializable (digital!) twin to the process table
// returned from discoveries. The processes (process tree) is represented in
// JSON as a JSON object, where the members (keys) are the stringified PIDs
// and the values are process objects.
type ProcessTable lxkns.ProcessTable

// MarshalJSON returns the JSON textual representation of a process table.
func (p ProcessTable) MarshalJSON() ([]byte, error) {
	// Similar to Golang's mapEncoder.encode, we iterate over the key-value
	// pairs ourselves, because we need to serialize alias types for the
	// individual process values, not the process values verbatim. By
	// iterating ourselves, we avoid building a new transient map with process
	// alias objects.
	b := bytes.Buffer{}
	b.WriteRune('{')
	first := true
	for _, proc := range p {
		if first {
			first = false
		} else {
			b.WriteRune(',')
		}
		b.WriteRune('"')
		b.WriteString(strconv.Itoa(int(proc.PID)))
		b.WriteString(`":`)
		procjson, err := json.Marshal(proc)
		if err != nil {
			return nil, err
		}
		b.Write(procjson)
	}
	b.WriteRune('}')
	return b.Bytes(), nil
}

// TypedNamespaceSet is the JSON representation a set of typed
// namespace IDs. The set is represented as a JSON object, with the keys being
// the namespace types and the IDs then being the number values.
type TypedNamespacesSet lxkns.NamespacesSet

func (n TypedNamespacesSet) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	b.WriteRune('{')
	first := true
	for idx, ns := range n {
		if ns == nil {
			continue
		}
		if first {
			first = false
		} else {
			b.WriteRune(',')
		}
		b.WriteRune('"')
		b.WriteString(lxkns.TypesByIndex[idx].Name())
		b.WriteString(`":`)
		nsjson, err := json.Marshal(ns.ID().Ino)
		if err != nil {
			return nil, err
		}
		b.Write(nsjson)
	}
	b.WriteRune('}')
	return b.Bytes(), nil
}

// FIXME: implement
func (n TypedNamespacesSet) UnmarshalJSON(data []byte) error {
	return nil
}
