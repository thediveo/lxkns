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

package plural

import (
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var printer = message.NewPrinter(language.English)

// Elements returns the specified count of elements together with the correct
// singular or plural form for elements, depending on the count. The elements
// parameter must specify the plural. One or more optional arguments can be
// specified for elements that possess varying elements, such as "hidden [TYP]
// namespaces".
func Elements(count int, elements string, a ...interface{}) string {
	return printer.Sprintf("%d "+elements, append([]interface{}{count}, a...)...)
}

type element struct {
	Singular string
	Plural   string
}

var elements = [...]element{
	{"bind-mounted namespace", "bind-mounted namespaces"},
	{"container", "containers"},
	{"container engine", "container engines"},
	{"composer project", "composer projects"},
	{"fd-referenced namespace", "fd-referenced namespaces"},
	{"hidden %s namespace", "hidden %s namespaces"},
	{"%s namespace", "%s namespaces"},
	{"mount point", "mount points"},
	{"pod", "pods"},
	{"process", "processes"},
}

func init() {
	for _, el := range elements {
		err := message.Set(language.English, "%d "+el.Plural,
			plural.Selectf(
				1, "%d",
				"=1", "%d "+el.Singular,
				"other", "%d "+el.Plural))
		if err != nil {
			panic(err)
		}
	}
}
