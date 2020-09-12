package types

import (
	"encoding/json"

	"github.com/thediveo/lxkns"
)

type DiscoveryResult lxkns.DiscoveryResult

func (d *DiscoveryResult) MarshalJSON() ([]byte, error) {
	nsdict := &NamespacesDict{
		AllNamespaces: &d.Namespaces,
		ProcessTable: ProcessTable{
			ProcessTable: d.Processes,
			Namespaces:   nil,
		},
	}
	nsdict.ProcessTable.Namespaces = nsdict
	aux := struct {
		Namespaces *NamespacesDict `json:"namespaces"`
		Processes  ProcessTable    `json:"processes"`
	}{
		Namespaces: nsdict,
		Processes:  nsdict.ProcessTable,
	}
	return json.Marshal(aux)
}
