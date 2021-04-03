// Copyright 2021 Harald Albrecht.
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

import { rgba } from './rgba'

describe('rgba', () => {

    it('parses different color formats and applies alpha', () => {
        expect(rgba('#123', 1)).to.equal('rgba(17,34,51,1)')
        expect(rgba('#112233', 0.5)).to.equal('rgba(17,34,51,0.5)')
        expect(rgba('rgb(1,2,3)', 0.5)).to.equal('rgba(1,2,3,0.5)')
        expect(rgba('red', 0.5)).to.equal('rgba(255,0,0,0.5)')
    })

    it('parses colors with alpha and correctly applies another alpha', () => {
        expect(rgba('rgba(1,2,3,.5)', 0.5)).to.equal('rgba(1,2,3,0.25)')
    })

})
