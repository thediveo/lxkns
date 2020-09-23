package output

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// NamespaceTypeIcons maps individual Linux-kernel namespace types
// (species.NNamespaceTypeIcons) to Unicode characters to be used as icons.
var NamespaceTypeIcons = map[species.NamespaceType]string{
	species.CLONE_NEWCGROUP: "🔧",
	species.CLONE_NEWIPC:    "✉ ",
	species.CLONE_NEWNS:     "📁",
	species.CLONE_NEWNET:    "⇄ ",
	species.CLONE_NEWPID:    "🏃",
	species.CLONE_NEWUSER:   "👤",
	species.CLONE_NEWUTS:    "💻",
}

// NamespaceIcon returns an Unicode string which can be displayed as an "icon"
// for the specified namespace. If showing namespace icons is disabled, then an
// empty string is always returned instead. If necessary, the returned string
// contains padding.
func NamespaceIcon(ns model.Namespace) (icon string) {
	if showNamespaceIcons {
		icon = NamespaceTypeIcons[ns.Type()] + " "
	}
	return
}

// showNamespaceIcons enables displaying Unicode "icons" along namespaces.
var showNamespaceIcons bool

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.RegisterPlugin(&plugger.PluginSpec{
		Name:  "icon",
		Group: "cli",
		Symbols: []plugger.Symbol{
			plugger.NamedSymbol{Name: "SetupCLI", Symbol: IconSetupCLI},
		},
	})
}

// IconSetupCLI is a plugin function that registers the CLI "--icon" flag.
func IconSetupCLI(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolVar(&showNamespaceIcons,
		"icon", false,
		"show/hide unicode icons next to namespaces")
}
