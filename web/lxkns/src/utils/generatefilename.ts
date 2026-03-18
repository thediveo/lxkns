// Copyright 2024 Harald Albrecht.
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

// Generates an at least slightly useful unique filename, given a base filename.
export const generateFilename = (basename: string, ext: string) => {
    const now = new Date()
    const d = now.getFullYear()
        + (now.getMonth() + 1).toString().padStart(2, '0')
        + now.getDay().toString().padStart(2, '0')
    const tod = now.getHours().toString().padStart(2, '0')
        + now.getMinutes().toString().padStart(2, '0')
        + now.getSeconds().toString().padStart(2, '0')
    return `${basename}-${d}-${tod}.${ext}`
}
