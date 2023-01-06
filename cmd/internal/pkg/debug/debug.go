package cli

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
	"github.com/thediveo/lxkns/log"

	_ "github.com/thediveo/lxkns/log/logrus"
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
	cmd.PersistentFlags().Bool("debug", false, "enables debug logging output")
}

func BeforeCommand(cmd *cobra.Command) error {
	if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
		log.SetLevel(log.DebugLevel)
		log.Debugf("debug logging enabled")
	} else {
		log.SetLevel(log.FatalLevel)
	}
	return nil
}
