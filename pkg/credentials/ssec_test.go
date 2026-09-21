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

package credentials

import (
	machineryapi "github.com/cloudnative-pg/machinery/pkg/api"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	barmanApi "github.com/cloudnative-pg/barman-cloud/pkg/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSE-C customer key", func() {
	It("is not read from its secret", func(ctx SpecContext) {
		// the key secret does not exist
		c := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
		configuration := &barmanApi.BarmanObjectStoreConfiguration{
			DestinationPath: "s3://bucket/path",
			BarmanCredentials: barmanApi.BarmanCredentials{AWS: &barmanApi.S3Credentials{
				InheritFromIAMRole: true,
				SSECustomerKey: &machineryapi.SecretKeySelector{
					LocalObjectReference: machineryapi.LocalObjectReference{Name: "missing-secret"},
					Key:                  "key",
				},
			}},
		}

		_, err := EnvSetCloudCredentialsAndCertificates(ctx, c, "default", configuration, nil, "")
		Expect(err).ToNot(HaveOccurred())
	})
})
