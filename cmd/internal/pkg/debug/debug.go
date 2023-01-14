package cli

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
	"github.com/thediveo/lxkns/log"

	_ "github.com/thediveo/lxkns/log/logrus"
)

// Names of the CLI flags defined and used in this package.
const (
	DebugFlagName = "debug"
	LogFlagName   = "log"
)

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		SetupCLI, plugger.WithPlugin("debug"))
	plugger.Group[cliplugin.BeforeCommand]().Register(
		BeforeCommand, plugger.WithPlugin("debug"))
}

// SetupCLI adds the "--debug" flag to the specified command that changes the
// logging level to debug.
func SetupCLI(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(DebugFlagName, false, "enables debug logging output")
	cmd.PersistentFlags().Bool(LogFlagName, false, "enables logging output (but no debug logging)")
}

// Ensure to enable debug logging before any command finally is executed.
func BeforeCommand(cmd *cobra.Command) error {
	if debug, _ := cmd.PersistentFlags().GetBool(DebugFlagName); debug {
		log.SetLevel(log.DebugLevel)
		log.Debugf("debug logging enabled")
	} else if logging, _ := cmd.PersistentFlags().GetBool(LogFlagName); logging {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.FatalLevel)
	}
	return nil
}
