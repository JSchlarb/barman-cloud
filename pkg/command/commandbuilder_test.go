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

package command

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"

	machineryapi "github.com/cloudnative-pg/machinery/pkg/api"

	barmanApi "github.com/cloudnative-pg/barman-cloud/pkg/api"
	"github.com/cloudnative-pg/barman-cloud/pkg/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("barmanCloudWalRestoreOptions", func() {
	var storageConf *barmanApi.BarmanObjectStoreConfiguration
	BeforeEach(func() {
		storageConf = &barmanApi.BarmanObjectStoreConfiguration{
			DestinationPath: "s3://bucket-name/",
		}
	})

	It("should generate correct arguments without the wal stanza", func(ctx SpecContext) {
		options, err := CloudWalRestoreOptions(ctx, storageConf, "test-cluster")
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.Join(options, " ")).
			To(
				Equal(
					"s3://bucket-name/ test-cluster",
				))
	})

	It("should generate correct arguments", func(ctx SpecContext) {
		extraOptions := []string{"--read-timeout=60", "-vv"}
		storageConf.Wal = &barmanApi.WalBackupConfiguration{
			RestoreAdditionalCommandArgs: extraOptions,
		}
		options, err := CloudWalRestoreOptions(ctx, storageConf, "test-cluster")
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.Join(options, " ")).
			To(
				Equal(
					"s3://bucket-name/ test-cluster --read-timeout=60 -vv",
				))
	})
})

var _ = Describe("useDefaultAzureCredentials", func() {
	It("should be false by default", func(ctx SpecContext) {
		Expect(useDefaultAzureCredentials(ctx)).To(BeFalse())
	})

	It("should be false if ctx contains an invalid value", func(ctx SpecContext) {
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, "invalidValue")
		Expect(useDefaultAzureCredentials(newCtx)).To(BeFalse())
	})

	It("should be false if ctx contains false value", func(ctx SpecContext) {
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, false)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeFalse())
	})

	It("should be true only if ctx contains true value", func(ctx SpecContext) {
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, true)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeTrue())
	})
})

var _ = Describe("AppendCloudProviderOptions with Azure credentials", func() {
	var options []string

	BeforeEach(func() {
		options = []string{}
	})

	It("should use default credential when UseDefaultAzureCredentials is set", func(ctx SpecContext) {
		credentials := barmanApi.BarmanCredentials{
			Azure: &barmanApi.AzureCredentials{
				UseDefaultAzureCredentials: true,
			},
		}
		result, err := appendCloudProviderOptions(ctx, options, credentials)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(ContainElements(
			"--cloud-provider", "azure-blob-storage",
			"--credential", "default",
		))
	})

	It("should use managed-identity credential when InheritFromAzureAD is set", func(ctx SpecContext) {
		credentials := barmanApi.BarmanCredentials{
			Azure: &barmanApi.AzureCredentials{
				InheritFromAzureAD: true,
			},
		}
		result, err := appendCloudProviderOptions(ctx, options, credentials)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(ContainElements(
			"--cloud-provider", "azure-blob-storage",
			"--credential", "managed-identity",
		))
	})

	It("should not use any credential flag for explicit credentials", func(ctx SpecContext) {
		credentials := barmanApi.BarmanCredentials{
			Azure: &barmanApi.AzureCredentials{
				StorageAccount: &machineryapi.SecretKeySelector{
					LocalObjectReference: machineryapi.LocalObjectReference{
						Name: "test",
					},
					Key: "account",
				},
				StorageKey: &machineryapi.SecretKeySelector{
					LocalObjectReference: machineryapi.LocalObjectReference{
						Name: "test",
					},
					Key: "key",
				},
			},
		}
		result, err := appendCloudProviderOptions(ctx, options, credentials)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal([]string{
			"--cloud-provider", "azure-blob-storage",
		}))
	})

	It("should use default credential from context when context flag is set", func(ctx SpecContext) {
		credentials := barmanApi.BarmanCredentials{
			Azure: &barmanApi.AzureCredentials{},
		}
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, true)
		result, err := appendCloudProviderOptions(newCtx, options, credentials)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(ContainElements(
			"--cloud-provider", "azure-blob-storage",
			"--credential", "default",
		))
	})

	It("should prioritize UseDefaultAzureCredentials over InheritFromAzureAD", func(ctx SpecContext) {
		credentials := barmanApi.BarmanCredentials{
			Azure: &barmanApi.AzureCredentials{
				UseDefaultAzureCredentials: true,
				InheritFromAzureAD:         true,
			},
		}
		result, err := appendCloudProviderOptions(ctx, options, credentials)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(ContainElements(
			"--cloud-provider", "azure-blob-storage",
			"--credential", "default",
		))
	})
})

var _ = Describe("AppendCloudProviderOptions with AWS credentials", func() {
	var options []string

	BeforeEach(func() {
		options = []string{}
	})

	It("should not add the SSE-C option when no customer key is set", func(ctx SpecContext) {
		credentials := barmanApi.BarmanCredentials{
			AWS: &barmanApi.S3Credentials{
				InheritFromIAMRole: true,
			},
		}
		result, err := appendCloudProviderOptions(ctx, options, credentials)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal([]string{
			"--cloud-provider", "aws-s3",
		}))
		Expect(result).ToNot(ContainElement("--sse-customer-key"))
	})

	It("should not add the SSE-C option to the cloud provider options", func(ctx SpecContext) {
		credentials := barmanApi.BarmanCredentials{
			AWS: &barmanApi.S3Credentials{
				InheritFromIAMRole: true,
				SSECustomerKey: &machineryapi.SecretKeySelector{
					LocalObjectReference: machineryapi.LocalObjectReference{
						Name: "sse-c-key",
					},
					Key: "key",
				},
			},
		}
		result, err := appendCloudProviderOptions(ctx, options, credentials)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal([]string{"--cloud-provider", "aws-s3"}))

		Expect(AppendSSECustomerKeyOption(nil, &barmanApi.BarmanObjectStoreConfiguration{
			BarmanCredentials: credentials,
		})).To(Equal([]string{
			"--sse-customer-key", "file://" + utils.SSECustomerKeyFilePath(credentials.AWS.SSECustomerKey),
		}))
	})
})

var _ = Describe("SSE-C customer key file per key reference", func() {
	credentialsWithKey := func(secretName, key string) barmanApi.BarmanCredentials {
		return barmanApi.BarmanCredentials{AWS: &barmanApi.S3Credentials{
			InheritFromIAMRole: true,
			SSECustomerKey: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{Name: secretName},
				Key:                  key,
			},
		}}
	}
	keyFileOption := func(_ SpecContext, credentials barmanApi.BarmanCredentials) string {
		options := AppendSSECustomerKeyOption(nil, &barmanApi.BarmanObjectStoreConfiguration{
			BarmanCredentials: credentials,
		})
		idx := slices.Index(options, "--sse-customer-key")
		Expect(idx).To(BeNumerically(">=", 0))
		return options[idx+1]
	}

	It("uses different key files for different secrets", func(ctx SpecContext) {
		Expect(keyFileOption(ctx, credentialsWithKey("key-a", "key"))).
			ToNot(Equal(keyFileOption(ctx, credentialsWithKey("key-b", "key"))))
	})

	It("uses different key files for different keys of the same secret", func(ctx SpecContext) {
		Expect(keyFileOption(ctx, credentialsWithKey("keys", "a"))).
			ToNot(Equal(keyFileOption(ctx, credentialsWithKey("keys", "b"))))
	})

	It("uses the same key file for the same key reference", func(ctx SpecContext) {
		Expect(keyFileOption(ctx, credentialsWithKey("key-a", "key"))).
			To(Equal(keyFileOption(ctx, credentialsWithKey("key-a", "key"))))
	})
})

var _ = Describe("SSE-C option of the barman-cloud commands", func() {
	var configuration *barmanApi.BarmanObjectStoreConfiguration
	var sseOption []string

	// fakeBarman puts an executable in PATH that records its arguments.
	fakeBarman := func(name, output string) string {
		dir := GinkgoT().TempDir()
		argsFile := filepath.Join(dir, name+".args")
		script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > " + argsFile + "\necho '" + output + "'\n"
		Expect(os.WriteFile(filepath.Join(dir, name), []byte(script), 0o700)).To(Succeed()) // #nosec G306
		GinkgoT().Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
		return argsFile
	}
	recordedArgs := func(argsFile string) []string {
		content, err := os.ReadFile(argsFile) // #nosec G304
		Expect(err).ToNot(HaveOccurred())
		return strings.Split(strings.TrimSpace(string(content)), "\n")
	}

	BeforeEach(func() {
		keyRef := &machineryapi.SecretKeySelector{
			LocalObjectReference: machineryapi.LocalObjectReference{Name: "sse-c-key"},
			Key:                  "key",
		}
		configuration = &barmanApi.BarmanObjectStoreConfiguration{
			DestinationPath: "s3://bucket/path",
			BarmanCredentials: barmanApi.BarmanCredentials{AWS: &barmanApi.S3Credentials{
				InheritFromIAMRole: true,
				SSECustomerKey:     keyRef,
			}},
		}
		sseOption = []string{"--sse-customer-key", "file://" + utils.SSECustomerKeyFilePath(keyRef)}
	})

	It("is passed to barman-cloud-wal-restore", func(ctx SpecContext) {
		options, err := CloudWalRestoreOptions(ctx, configuration, "cluster")
		Expect(err).ToNot(HaveOccurred())
		Expect(options).To(ContainElements(sseOption))
	})

	It("is passed to barman-cloud-backup-delete", func(ctx SpecContext) {
		argsFile := fakeBarman(utils.BarmanCloudBackupDelete, "")
		Expect(DeleteBackupsByPolicy(ctx, configuration, "cluster", nil, "7d")).To(Succeed())
		Expect(recordedArgs(argsFile)).To(ContainElements(sseOption))
	})

	It("is passed to barman-cloud-backup-list", func(ctx SpecContext) {
		argsFile := fakeBarman(utils.BarmanCloudBackupList, `{"backups_list": []}`)
		_, err := GetBackupList(ctx, configuration, "cluster", nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(recordedArgs(argsFile)).To(ContainElements(sseOption))
	})

	It("is passed to barman-cloud-backup-show", func(ctx SpecContext) {
		argsFile := fakeBarman(utils.BarmanCloudBackupShow, `{"cloud": {}}`)
		_, err := GetBackupByName(ctx, "backup", "cluster", configuration, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(recordedArgs(argsFile)).To(ContainElements(sseOption))
	})

	It("is not added when no SSE-C key is configured", func(ctx SpecContext) {
		configuration.AWS.SSECustomerKey = nil
		options, err := CloudWalRestoreOptions(ctx, configuration, "cluster")
		Expect(err).ToNot(HaveOccurred())
		Expect(options).ToNot(ContainElement("--sse-customer-key"))
	})
})
