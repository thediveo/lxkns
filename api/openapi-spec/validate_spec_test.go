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

package openapispec

import (
	"context"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("lxkns OpenAPI specification", func() {

	It("validates", func() {
		oas, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile("lxkns.yaml")
		Expect(err).To(Succeed())
		Expect(oas.Validate(context.WithTimeout(context.Background(), 10*time.Second))).To(Succeed())

	})

})
