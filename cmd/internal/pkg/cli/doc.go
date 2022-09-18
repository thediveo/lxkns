/*
Package cli handles registering CLI flags via a plug-in mechanism. Additionally,
optional CLI flag handling can be carried out after the CLI flags have been
parsed and just right before the selected command is about to run.

Packages providing CLI flags need to register themselves using the go-plugger
mechanism in the plugin group named "cli". It is possible to register multiple
flags from the same package in a highly modular fashion just by specifying
individual plug-in names, for instance, per each individual flag. See the style
package for a working example.

Plug-ins in the "cli" group should export these functions:

  - SetupCLI: a "func(*cobra.Command)" registering CLI one or more flags.
  - BeforeRun: an optional "func() error" which is called before the command runs.

Registration of the exported functions should be done, as usual, in an init()
function. For better modularity, multiple such registration-related init()
functions can perfectly co-exist within the same package. Just make sure to
specify different plug-in names.

It already sufficies when a cmd package references an CLI-related package to
pull in its CLI flag registrations. A cmd package then should make sure to call
cli.AddFlags() and cli.BeforeCommand() respectively.

AddFlags() should be called after your cmd package has created the root command
object and is ready for registering flags.

BeforeCommand() should be called from the PersistentPreRunE hook function of
your cmd package's root command object.
*/
package cli
