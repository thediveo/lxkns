# mntnssandbox

`mntnssandbox` is an *optional* binary that can be built from
`cmd/mntnssandbox`. It does nothing more than upon start to immediately switch
into the ("target") mount namespace passed to it via the `sleepy_mntns`
environment variable, saying "OK", and then go to sleep.

If available, `mntnssandbox` is automatically used by the
[mountineers](mountineers) API when in need of a "sandbox" process.

The "OK" is a necessary synchronization message as the process spinning up the
sandbox process otherwise cannot know when the sandbox process has switched its
mount namespace. Until the "OK" we cannot reliably access the target mount
namespace via the sandbox process' `/proc/[PID]/root/` entry.

Optionally, the `sleepy_userns` specifies the user namespace to enter first
before entering the mount namespace. This covers the case where the process
hasn't the required capabilities for switching into the mount namespace in the
current user namespace but will gain them by switching into the specified user
namespace first.

If the lxkns service or any application including lxkns finds the `mntnssandbox`
binary in its `PATH` then it'll automatically use it, otherwise it falls back to
re-executing itself and immediately putting the copy to sleep. The advantage of
the separate `mntnssandbox` binary is that it uses less system memory resources
for sleeping: the dedicated `mntnssandbox` binary is much smaller so only needs
a small RAM "bed" for sleeping, compared to a usually much larger application
using lxkns.