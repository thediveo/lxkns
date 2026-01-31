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

import { useState } from 'react'

// We simply create a series of id numbers, whenenver one is requested. Using
// the state hook we then ensure that a functional component gets its own stable
// id, yet multiple components get their own stable individual ids. Why 42? Oh,
// read the pop classics!
let someId = 42

// Returns a unique id using the given prefix or default of 'id-'.
const useId = (prefix = 'id-') => {
    // Only calculate a new id if useState really needs an initial state;
    // otherwise, skip it to keep things stable over the livetime of a single
    // component.
    const [ id ] = useState(() => prefix + (someId++).toString())
    return id
};

export default useId
