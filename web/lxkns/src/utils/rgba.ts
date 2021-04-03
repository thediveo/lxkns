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

import colorRgba from 'color-rgba'

/**
 * Returns a CSS "rgba(...)" color string given a CSS color string (which
 * optionally might include an alpha value itself) and a separate alpha
 * (transparency) value.
 *
 * @param color color string, such as "#rgb", "#rrggbb", "rgb(r,g,b)", et
 * cetera. Even "rgba(r,g,b,a)" is acceptable.
 * @param alpha alpha value in the range of [0..1].
 */
export const rgba = (color: string, alpha: number) => {
    const [r, g, b, a] = colorRgba(color)
    return `rgba(${r},${g},${b},${a*alpha})`
}
