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

import { atom, Getter, Setter } from 'jotai'


/**
 * Given a key, returns the corresponding setting from local storage, or failing
 * to find any useful stored setting, returns a default setting specified by the
 * caller.
 *
 * The type of the storage setting must match the type of the default value,
 * otherwise the default value will be returned instead.
 *
 * In case of boolean settings any string value will be automatically converted
 * into a boolean value first.
 *
 * @param storageKey name/identifier of local storage item.
 * @param defaultValue default boolean value in case no (valid) local storage
 * data could be found for this setting.
 */
const initialAtomValue = <T>(storageKey: string, defaultValue: T): T => {
    try {
        let json = localStorage.getItem(storageKey)
        if (json !== null) {
            var value
            try {
                value = JSON.parse(json)
            } catch (e) {
                // The stored setting isn't valid a JSON value, so take it as a
                // verbatim string.
                value = json
            }
            // If the setting is a boolean, then we additionally accept a stored
            // string which we then map onto boolean.
            if (typeof(defaultValue) === 'boolean') {
                if (typeof(value) === 'string') {
                    if (value === 'on' || 'true'.startsWith(value.toLowerCase())) {
                        return true as unknown as T
                    } else if (value === 'off' || 'false'.startsWith(value.toLowerCase())) {
                        return false as unknown as T
                    }
                    return defaultValue
                }
            }
            // Otherwise, the type of the stored setting must match the type of
            // the default value.
            return typeof(value) === typeof(defaultValue) ? value : defaultValue
        }
    } catch (e) { }
    return defaultValue
}

/**
 * Returns a new atom storing settings state, identified by the given key in
 * (browser) local storage.
 *
 * @param storageKey name of local storage item.
 * @param defaultValue default value in case no local storage data could be
 * found for this setting. Please note that when the default value is taken,
 * then it won't be written back to the local storage item, but instead the
 * non-existing storage item will be left non-existing.
 */
export const localStorageAtom = <T>(storageKey: string, defaultValue: T) => {
    const storageAtom = atom(
        initialAtomValue(storageKey, defaultValue),
        (_get: Getter, set: Setter, arg: T) => {
            set(storageAtom, arg)
            localStorage.setItem(storageKey, JSON.stringify(arg))
        }
    )
    return storageAtom
}
