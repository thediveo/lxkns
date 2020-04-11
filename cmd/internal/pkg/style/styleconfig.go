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

	"gopkg.in/yaml.v2"
)

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

// Parse a single element style, such as a color settings, or a styling
// attribute like "bold", et cetera.
func parseElementStyle(sty *Style, elementStyle interface{}) {
	// If it's "just" a simple string (list) element, then we interpret it as
	// an attribute, which must be one of the defined styling attributes.
	if attr, ok := elementStyle.(string); ok {
		switch attr {
		case "blink": // AAAARGH!!!!
			sty.style = sty.style.Blink()
		case "bold":
			sty.style = sty.style.Bold()
		case "crossout":
			sty.style = sty.style.CrossOut()
		case "faint":
			sty.style = sty.style.Faint()
		case "italic", "italics":
			sty.style = sty.style.Italic()
		case "overline":
			sty.style = sty.style.Overline()
		case "reverse":
			sty.style = sty.style.Reverse()
		case "underline":
			sty.style = sty.style.Underline()
		default:
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
