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

package lxkns

import (
	"encoding/json"
	"os"
	"os/user"
	"strconv"

	"github.com/thediveo/gons/reexec"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
)

// UidUsernameMap maps user identifiers (uids) to their corresponding user
// names, if any.
type UidUsernameMap map[uint32]string

// A list of uids to ask the user information oracle for.
type userIds []uint32

// DiscoverUserNames returns the mapping from user identifiers (uids, found as
// owners of user namespaces) to their corresponding user names, if any.
func DiscoverUserNames(namespaces model.AllNamespaces) UidUsernameMap {
	useridmap := UidUsernameMap{} // not the real one, but just for collecting unique uids
	// Scan all user namespaces for their owner uids; all other namespaces
	// don't have owning users, but instead owning user namespaces.
	for _, userns := range namespaces[model.UserNS] {
		owneruid := uint32(userns.(model.Ownership).UID())
		if _, ok := useridmap[owneruid]; !ok {
			useridmap[owneruid] = ""
		}
	}
	uids := userIds{}
	for uid := range useridmap {
		uids = append(uids, uid)
	}
	return queryUserNames(uids, namespaces)
}

// queryUserNames is a lower-level function that takes a set of uids and
// queries the operating system, swiching into the initial mount namespace if
// necessary.
//
// Note: /etc/passwd actually isn't the single point of truth. Normally,
// getpwnam() (getpwuid(), ...) handles the situation of having multiple
// points of truth for user information. But they are designed to be oracles,
// in the sense that only given a concrete uid or user name they will answer
// with concrete information about this user, and only this particular user.
// So we need to pass in a list of uids for which we're seeking the
// corresponding user names.
func queryUserNames(uids userIds, namespaces model.AllNamespaces) UidUsernameMap {
	var usernames UidUsernameMap
	jsonuids, err := json.Marshal(uids)
	if err != nil {
		return usernames
	}
	envvars := []string{"UIDS=" + string(jsonuids)}
	// We need to read the user names while in the initial mount namespace, as
	// otherwise we'll end up with the wrong /etc/passwd. If we cannot access
	// the initial mount namespace, then silently fall back to reading from
	// our current mount namespace.
	mntnsid, err := ops.NamespacePath("/proc/1/ns/mnt").ID()
	if err != nil {
		log.Infof("cannt access initial mount namespace, falling back to own mount namespace")
		return userNamesOracle(uids)
	}
	if err := ReexecIntoActionEnv(
		"discover-uid-names",
		MountEnterNamespaces(namespaces[model.MountNS][mntnsid], namespaces),
		envvars, &usernames); err != nil {
		log.Errorf("cannot read user name information from initial mount namespace")
		return UidUsernameMap{}
	}
	return usernames
}

// Register re-execution action for reading the user names for a set of given
// uids.
func init() {
	reexec.Register("discover-uid-names", readUidNames)
}

// readUidNames handles re-execution as an action: it gathers the uids to look
// up which get passed via an environment variable, runs the query, and then
// sends back the results as JSON via stdout.
func readUidNames() {
	uidsparam, ok := os.LookupEnv("UIDS")
	if !ok {
		panic("missing uids")
	}
	var uids userIds
	if err := json.Unmarshal([]byte(uidsparam), &uids); err != nil {
		panic(err.Error())
	}
	if err := json.NewEncoder(os.Stdout).Encode(userNamesOracle(uids)); err != nil {
		panic(err.Error())
	}
}

// userNamesOracle looks up the (loginq) names for the given user identifiers.
func userNamesOracle(uids userIds) UidUsernameMap {
	usernames := UidUsernameMap{}
	for _, uid := range uids {
		if u, _ := user.LookupId(strconv.FormatUint(uint64(uid), 10)); u != nil {
			usernames[uid] = u.Username
		}
	}
	return usernames
}
