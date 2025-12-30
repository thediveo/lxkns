package turtles

import (
	"context"
	"time"

	"github.com/siemens/turtlefinder"
	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/model"
)

// Names of the CLI flags defined and used in this package.
const (
	WaitFlagName         = "wait"
	NoContainersFlagName = "no-containers"
)

// Containerizer returns a TurtleFinder containerizer that autodetects the
// available container engines. The containerizer is set up to wait a CLI flag
// specified time for synchronizing to the workload of newly found container
// engines; please note that this wait happens when discovering containers, not
// when creating the containerizer itself.
func Containerizer(ctx context.Context, cmd *cobra.Command) containerizer.Containerizer {
	if noContainers, _ := cmd.PersistentFlags().GetBool(NoContainersFlagName); noContainers {
		return &NullContainerizer{}
	}

	maxwait, _ := cmd.PersistentFlags().GetDuration(WaitFlagName)
	tf := turtlefinder.New(
		func() context.Context { return ctx },
		turtlefinder.WithGettingOnlineWait(maxwait),
	)
	return tf
}

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		setupCLI, plugger.WithPlugin("lxkns/turtles"))
}

// setupCLI registers the engine-agnostic specific CLI options.
func setupCLI(cmd *cobra.Command) {
	pf := cmd.PersistentFlags()
	pf.Duration(WaitFlagName, 3*time.Second,
		"max duration to wait for container engine workload synchronization before continuing")
	pf.Bool(NoContainersFlagName, false, "skip container discovery")
}

type NullContainerizer struct{}

var _ containerizer.Containerizer = (*NullContainerizer)(nil)

func (n *NullContainerizer) Containers(ctx context.Context, procs model.ProcessTable, pidmap model.PIDMapper) []*model.Container {
	return nil
}

func (n *NullContainerizer) Close() {}
