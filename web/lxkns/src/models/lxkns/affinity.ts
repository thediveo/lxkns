// Copyright 2026 Harald Albrecht.
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

/**
 * Returns true if two CPU affinity lists are quivalent, vulgo: "the same".
 * 
 * @param cpusA first CPU affinity list, or null.
 * @param cpusB second CPU affinity list, or null.
 * @returns true only if both CPU affinity lists are equivalent.
 */
export const sameAffinity = (cpusA: number[][] | null, cpusB: number[][] | null) => {
    if (!cpusA || !cpusB) return false
    if (cpusA.length != cpusB.length) return false
    return !cpusA.some(([fromA, toA], idx) => {
        const [fromB, toB] = cpusB[idx]
        return fromA != fromB || toA != toB
    })
}

/**
 * Returns the number of (logical) CPUs mentioned in an affinity list.
 * 
 * @param cpus CPU affinity list, or null
 * @returns number of CPUs mentioned in affinity list, or 0 if empty or null.
 */
export const numCPUs = (cpus: number[][] | null) =>
    cpus?.reduce((sum, [from, to]) => (to - from + 1) + sum, 0) || 0
