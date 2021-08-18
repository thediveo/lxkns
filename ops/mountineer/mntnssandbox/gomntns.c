/*
 * Initializer function to join this(!) process to a specific Linux-kernel mount
 * namespace before the Go runtime spins up and blocks switching the mount
 * namespace, especially mount namespaces due to creating multiple OS threads.
 *
 * This initializer DOES NOT return, but instead blocks forever.
 *
 * Copyright 2019, 2021 Harald Albrecht.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License.You may obtain a copy of
 * the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations under
 * the License.
 */

/* Fun stuff... */
#define _GNU_SOURCE
#include <sched.h>
#include <unistd.h>
#include <sys/syscall.h>

/* Booooring stuff... */
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <fcntl.h>
#include <stdarg.h>
#include <errno.h>
#include <limits.h>

#define MNT_ENVVAR "sleepy_mntns"
#define USER_ENVVAR "sleepy_userns"

/*
 * Switch into the Linux kernel mount namespace specified through an env
 * variable. This env var references the namespace via a filesystem path. If
 * this env var is unspecified, there's nothing to switch. Optionally, switch
 * the user namespace first.
 *
 * After switching (or an attempted) send a message and then block indefinitely.
 * If no switch has been requested, then it silently returns to the caller.
 */
void gosandbox(void) {
    // Do we need to switch the user namespace first?
    char *usernsref = getenv(USER_ENVVAR);
    if (usernsref && *usernsref) {
        int usernsfd = open(usernsref, O_RDONLY);
        if (usernsfd < 0) {
            dprintf(STDERR_FILENO, 
                "package mntnssandbox: invalid user namespace reference \"%s\": %s\n", 
                usernsref, strerror(errno));
            exit(66);
        }
        /*
         * Do not use the glibc version of setns, but go for the syscall itself.
         * This allows us to avoid dynamically linking to glibc even when using
         * cgo, resorting to musl, et cetera. As musl is a mixed bag in terms of
         * its glibc compatibility, especially in such dark corners as Linux
         * namespaces, we try to minimize potentially problematic dependencies
         * here.
         *
         * A useful reference is Dominik Honnef's blog post "Statically compiled
         * Go programs, always, even with cgo, using musl":
         * https://dominik.honnef.co/posts/2015/06/statically_compiled_go_programs__always__even_with_cgo__using_musl/
         */
        long res = syscall(SYS_setns, usernsfd, CLONE_NEWUSER);
        close(usernsfd); /* Don't leak file descriptors */
        if (res < 0) {
            dprintf(STDERR_FILENO,
                "package gons: cannot join user namespace using reference \"%s\": %s\n", 
                usernsref, strerror(errno));
            exit(66);
        }
    }
    // And now let's switch the mount namespace.
    char *mntnsref = getenv(MNT_ENVVAR);
    if (!mntnsref || !*mntnsref) {
        return; // proceed
    }
    // Try to reference the specified mount namespace and then switch.
    int mntnsfd = open(mntnsref, O_RDONLY);
    if (mntnsfd < 0) {
        dprintf(STDERR_FILENO, 
            "package mntnssandbox: invalid mount namespace reference \"%s\": %s\n", 
            mntnsref, strerror(errno));
        exit(66);
    }
    long res = syscall(SYS_setns, mntnsfd, CLONE_NEWNS);
    close(mntnsfd); /* Don't leak file descriptors */
    if (res < 0) {
        dprintf(STDERR_FILENO,
            "package gons: cannot join mount namespace using reference \"%s\": %s\n", 
            mntnsref, strerror(errno));
        exit(66);
    }
    // Work suckzessfully done. Sleep.
    dprintf(STDOUT_FILENO, "OK\n");
    for (;;) {
        pause(); // pause might get interrupted, so wrap it.
    }
}
