# CPU Affinities View

This "Core Fancy" view show for all currently online logical CPUs the processes
and tasks that are allowed to be executed on them. This view is particular
useful when working with so-called "CPU isolation", where only a particular
process or set of processes are allowed to execute on an "isolated" CPU. A
typical example might be industrial automation control, where timing-sensitive
control processes must meet strict scheduling deadlines, otherwise this results
in expensive machinery stops or even damage.

> [!WARNING] On hosts with "a lot" of logical CPUs browsers start to struggle
> with rendering large or many subtrees.

This view lists all currently online logical CPUs ➊ by their numbers. Below each
logical CPU are the roots ➋ of the two process/task hierarchies:

- **PID&nbsp;2** for the kernel-space processes, which are strictly termed
  "kernel threads" despite appearing as processes.
- **PID&nbsp;1** for all the user-space processes and tasks (~threads).

![view CPU affinities and realtime scheduling](_images/lxkns-core-fancy-view.png ':class=framedscreenshot')

> [!ATTENTION] These hierarchies also include processes that are **not** allowed
> to execute on this particular CPU; this is to show the context of those
> processes that are "pinned" to this particular CPU.

- the **information about an process is grayed out completely** when it is not
  allowed to run on the particular CPU a process hierarchy is shown for.
- the **information about a process is semi-transparent** when it is allowed to
  run on this particular CPU, but not specifically restricted ("pinned") to it.

Interestingly, the codespace VM has a
[`multipathd`](https://manpages.org/multipathd/8) process ➎ that is running with
realtime scheduling: when it needs to run it cannot be interrupted anymore as
its scheduling policy is FIFO (first in, first out) at the highest priority 99
of the Linux kernel. `multipathd` handles multiple paths to a storage device.

### Notes

[^PID1]: as particular exemplars of PID&nbsp;1 consider themselves to be
    "L'État, c'est moi" we're adorning them always with a nice fake-golden
    crown. Or maybe, it's orange?
