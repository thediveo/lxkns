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
	"os"
	"reflect"

	"github.com/muesli/termenv"
	"gopkg.in/yaml.v2"
)

// parseStyles parses a YAML string containing style information into the
// package styles, such as MntStyle, et cetera.
func parseStyles(configyaml string) {
	yml := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(configyaml), &yml)
	if err != nil {
		fmt.Fprint(os.Stderr,
			"error: failed to parse style configuration yaml; "+
				"is this valid yaml?\n")
		return
	}
	// The top-level elements reference elements which can be styled. If we
	// find unknown elements, then we just print a warning and then carry on.
	for elementKey, elementStyles := range yml {
		sty, ok := Styles[elementKey]
		if !ok {
			fmt.Fprintf(os.Stderr,
				"warning: unknown element %q\n", elementKey)
			continue
		}
		// Below the top-level stylable elements come the individual styles
		// which can be configured. There are two types of styles: colors and
		// attributes (such as bold, italics, et cetera). Colors are expressed
		// as "color: colorcode" elements, while attributes are just plain
		// "attribute" elements.
		for _, elementStyle := range elementStyles.([]interface{}) {
			parseElementStyle(sty, elementStyle)
		}
	}
}

// Maps style attributes to their corresponding termenv.Style methods.
var styleAttributeMap = map[string]string{
	"blink":     "Blink",
	"bold":      "Bold",
	"crossout":  "CrossOut",
	"faint":     "Faint",
	"italic":    "Italic",
	"italics":   "Italic",
	"overline":  "Overline",
	"reverse":   "Reverse",
	"underline": "Underline",
}

// Parse a single element style, such as a color settings, or a styling
// attribute like "bold", et cetera.
func parseElementStyle(sty *Style, elementStyle interface{}) {
	// If it's "just" a simple string (list) element, then we interpret it as
	// an attribute, which must be one of the defined styling attributes.
	if attr, ok := elementStyle.(string); ok {
		if stylemethod, ok := styleAttributeMap[attr]; ok {
			// You can call me by name ... *plonk*
			//
			// Fun fact: coming up with the reflection-based call reduces the
			// cyclomatic complexity significantly, while raising the idiocity
			// complexity by several orders of magnitude. We can call this a
			// clear win for gocyclo.
			sty.style = reflect.ValueOf(&sty.style).MethodByName(stylemethod).
				Call([]reflect.Value{})[0].Interface().(termenv.Style)
		} else {
			fmt.Fprintf(os.Stderr,
				"warning: unknown styling attribute %q\n", attr)
		}
	} else {
		// Otherwise it has to be an object representing a color mapping for
		// the foreground and/or background colors.
		if colormap, ok := elementStyle.(map[interface{}]interface{}); ok {
			for colorkey, color := range colormap {
				colorname, ok1 := colorkey.(string)
				colorvalue, ok2 := color.(string)
				if !ok1 || !ok2 {
					fmt.Fprintf(os.Stderr,
						"warning: unknown color %s: %q\n", colorname, colorvalue)
					continue
				}
				switch colorname {
				case "foreground":
					sty.style = sty.style.Foreground(colorProfile.Color(colorvalue))
				case "background":
					sty.style = sty.style.Background(colorProfile.Color(colorvalue))
				}
			}
		}
	}
}
