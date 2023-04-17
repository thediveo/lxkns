/*
Package nsioctl defines namespace-related ioctl request values that aren't
defined in the sys/unix standard package.

# References

  - See also [ioctl_ns(2)] for details about the namespace-related ioctl
    operations.
  - For background information on getting the network namespace of a TAP/TUN
    netdev please refer to [TUNGETDEVNETNS].

[ioctl_ns(2)]: https://man7.org/linux/man-pages/man2/ioctl_ns.2.html
[TUNGETDEVNETNS]: https://unix.stackexchange.com/a/743003
*/
package nsioctl
