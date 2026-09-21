/*
Copyright © contributors to CloudNativePG, established as
CloudNativePG a Series of LF Projects, LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"path"

	machineryapi "github.com/cloudnative-pg/machinery/pkg/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSE-C customer key file path", func() {
	ref := func(name, key string) *machineryapi.SecretKeySelector {
		return &machineryapi.SecretKeySelector{
			LocalObjectReference: machineryapi.LocalObjectReference{Name: name},
			Key:                  key,
		}
	}

	It("is <SSECustomerKeyDirectory>/<secret name>/<key>", func() {
		Expect(SSECustomerKeyFilePath(ref("sse-c-key", "key"))).
			To(Equal(path.Join(SSECustomerKeyDirectory, "sse-c-key", "key")))
	})

	// /controller is shared with the postgres container
	It("is outside the /controller directory", func() {
		Expect(SSECustomerKeyFilePath(ref("sse-c-key", "key"))).ToNot(HavePrefix("/controller/"))
	})

	It("differs for different references", func() {
		Expect(SSECustomerKeyFilePath(ref("a", "key"))).ToNot(Equal(SSECustomerKeyFilePath(ref("b", "key"))))
		Expect(SSECustomerKeyFilePath(ref("keys", "a"))).ToNot(Equal(SSECustomerKeyFilePath(ref("keys", "b"))))
	})
})
