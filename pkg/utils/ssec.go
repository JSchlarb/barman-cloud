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
)

// SSECustomerKeyDirectory is where the caller mounts the SSE-C keys.
const SSECustomerKeyDirectory = "/sse-customer-keys"

// SSECustomerKeyFilePath returns the key file path: one directory per secret,
// one file per key.
func SSECustomerKeyFilePath(keyRef *machineryapi.SecretKeySelector) string {
	return path.Join(SSECustomerKeyDirectory, keyRef.Name, keyRef.Key)
}
