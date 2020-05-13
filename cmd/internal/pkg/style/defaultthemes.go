// Defines the two namespace colorization themes, dark and light.

// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package style

const defaultDarkTheme = `# dark lxkns colorization theme
user:
- bold
- background: '#404040'
pid:
- bold
- background: '#c0ffff'
- foreground: '#000000'
cgroup:
- background: '#800000'
ipc:
- background: '#808000'
mnt:
- background: '#000080'
net:
- background: '#008000'
time:
- background: '#804000'
uts:
- background: '#800080'

process:
- foreground: '#00c000'
owner:
- foreground: '#e0e000'
unknown:
- foreground: '#ff0000'
`

const defaultLightTheme = `# light lxkns colorization theme

# types of kernel namespaces
user:
- bold
- background: '#dadada' # Schwitters would be delighted!
- foreground: '#000000'
pid:
- bold
- background: '#e6ffff'
- foreground: '#000000'
cgroup:
- background: '#ffe6e6'
ipc:
- background: '#ffffcc'
mnt:
- background: '#e6e6ff'
net:
- background: '#ccffdd'
time:
- background: '#ffcc99'
uts:
- background: '#d9b3ff'

# namespace-related elements
process:
- foreground: '#004000'
owner:
- foreground: '#808000'
unknown:
- foreground: '#800000'

# process capabilities in namespaces
user-nocaps:
- background: '#800000'
- foreground: '#ffffff'
user-effcaps:
- background: '#808000'
- foreground: '#ffffff'
user-fullcaps:
- background: '#008000'
- foreground: '#ffffff'
`
