package output

import (
	"github.com/thediveo/lxkns/nstypes"
)

var NamespaceTypeIcons = map[nstypes.NamespaceType]string{
	nstypes.CLONE_NEWCGROUP: "🔧",
	nstypes.CLONE_NEWIPC:    "✉",
	nstypes.CLONE_NEWNS:     "📁",
	nstypes.CLONE_NEWNET:    "⇄",
	nstypes.CLONE_NEWPID:    "🏃",
	nstypes.CLONE_NEWUSER:   "👤",
	nstypes.CLONE_NEWUTS:    "💻",
}
