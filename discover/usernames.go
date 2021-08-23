// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package discover

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/ops/mountineer"
)

// etcpasswd is the pathname of the password file.
const etcpasswd = "/etc/passwd"

// UidUsernameMap maps user identifiers (uids) to their corresponding user
// names, if any.
type UidUsernameMap map[uint32]string

// DiscoverUserNames returns the mapping from user identifiers (uids, found as
// owners of user namespaces) to their corresponding user names, if any. The
// namespaces information is required so that the information can be
// discovered from the initial mount namespace of the host.
func DiscoverUserNames(namespaces model.AllNamespaces) UidUsernameMap {
	// We need to read the user names while in the initial mount namespace, as
	// otherwise we'll end up with the wrong /etc/passwd. If we cannot access
	// the initial mount namespace, then silently fall back to reading from
	// our current mount namespace.
	mntnsid, err := ops.NamespacePath("/proc/1/ns/mnt").ID()
	if err != nil {
		log.Infof("cannot access initial mount namespace, falling back to own mount namespace")
		return userNamesFromPasswd(etcpasswd)
	}
	mymntnsid, err := ops.NamespacePath("/proc/self/ns/mnt").ID()
	if err == nil && mymntnsid == mntnsid {
		return userNamesFromPasswd(etcpasswd)
	}
	// Safety net: if we don't have information about the mount namespace of
	// process PID 1, then there's something rotten and we go for an empty
	// mapping instead.
	if namespaces[model.MountNS][mntnsid] == nil {
		log.Warnf("missing information about PID 1 mount namespace")
		return UidUsernameMap{}
	}
	mnteer, err := mountineer.NewWithMountNamespace(namespaces[model.MountNS][mntnsid], namespaces[model.UserNS])
	if err != nil {
		log.Errorf("cannot open mount namespace for VFS operations: %s", err.Error())
		return UidUsernameMap{}
	}
	defer mnteer.Close()
	etcpasswd, err := mnteer.Resolve(etcpasswd)
	if err != nil {
		log.Errorf("cannot translate path /etc/passwd in target mount namespace: %s", err.Error())
		return UidUsernameMap{}
	}
	return userNamesFromPasswd(etcpasswd)
}

// userNamesFromPasswd parses /etc/passwd in the currently active mount
// namespace and return the mapping from user IDs (uids) to user names.
func userNamesFromPasswd(passwdpath string) UidUsernameMap {
	usernames := UidUsernameMap{}
	f, err := os.Open(passwdpath)
	if err != nil {
		return usernames
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Scanning follows the rules the glibc seems to follow, ignoring
		// empty lines and lines starting with a "#" comment symbol.
		// Additionally, we skip user names which start with either "+" or
		// "-".
		line := scanner.Text()
		if line == "" || line[0] == '#' {
			continue
		}
		fields := strings.SplitN(line, ":", 4)
		if len(fields) < 4 || fields[0] == "" || fields[2] == "" ||
			fields[0][0] == '+' || fields[0][0] == '-' {
			continue
		}
		if uid, err := strconv.Atoi(fields[2]); err == nil {
			usernames[uint32(uid)] = fields[0]
		}
	}
	return usernames
}
