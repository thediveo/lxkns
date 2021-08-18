// Copyright 2021 Harald Albrecht.
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

package mntnssandbox

/*
extern void gosandbox(void);
void __attribute__((constructor)) mntnssandboxinit(void) {
	gosandbox();
}
*/
import "C"

const MntnsEnvironmentVariableName = "sleepy_mntns"
const UsernsEnvironmentVariableName = "sleepy_userns"

// There's nothing else here; gomntns will never return.
