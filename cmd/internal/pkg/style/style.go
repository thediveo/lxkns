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

import (
	"fmt"
	"strconv"

	"github.com/muesli/termenv"
)

// Style represents a set of style settings to apply to text or value when
// rendering to a terminal supporting ANSI coloring and styling. Style
// settings are foreground and background colors, bold, italic, underlined, et
// cetera.
type Style struct {
	style termenv.Style // the encapsulated termenv styling information.
}

// V returns the value in the given style, so it can later be formatted by
// common formatters, such as fmt.Printf, et cetera.
func (st Style) V(value interface{}) StyledValue {
	return StyledValue{value: value, style: st.style}
}

// S returns the specified text s styled according to this Style's
// configuration. If multiple strings are specified, then the styling is
// applied to each string anew, thus allowing interleaving differently styled
// strings with this styling. The individual strings are put immediately
// adjacent to each without any intermediate spaces.
func (st *Style) S(args ...interface{}) (s string) {
	for _, arg := range args {
		s += st.style.Styled(fmt.Sprint(arg))
	}
	return
}

// StyledValue represents a styled value which can be formatted in different
// formats, such as decimal, string, quoted string, et cetera.
type StyledValue struct {
	value interface{}   // the styled value
	style termenv.Style // the style of the value (not: format)
}

// String returns the styled text representation of the styled value.
func (sv StyledValue) String() string {
	return sv.style.Styled(fmt.Sprint(sv.value))
}

// Format is a custom formatter which formats a styled value. It does not
// introduce its own % formats, but instead relies on fmt.Sprintf for % format
// support and then styles the formatted value string. This was inspired by
// github.com/ogrusorgru/aurora's value custom formatter implementation.
func (sv StyledValue) Format(s fmt.State, c rune) {
	format := "%"
	if width, ok := s.Width(); ok {
		format += strconv.Itoa(width)
	}
	if precision, ok := s.Precision(); ok {
		format += "." + strconv.Itoa(precision)
	}
	format += string(c)
	s.Write([]byte(sv.style.Styled(fmt.Sprintf(format, sv.value))))
}
