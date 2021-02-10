// Copyright 2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//+build integration

package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAWSProviderValidityDaysExceedC(t *testing.T) {
	rootCAIdentifier := os.Getenv("ROOT_CA_ARN")
	templateArn := os.Getenv("TEMPLATE_ARN")

	cmd := exec.Command("../getistio", "gen-ca", "--provider=aws",
		fmt.Sprintf("--signing-ca=%s", rootCAIdentifier),
		fmt.Sprintf("--template-arn=%s", templateArn),
		"--validity-days=100000",
		"--secret-file-path=/tmp/file1.yaml")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	cmd.Stdout = os.Stdout
	cmd.Run()
	actual := buf.String()
	assert.Contains(t, actual, "unable to issue CA, due to error: unable to issue certificate: unable to issue certificate: ValidationException: The certificate validity specified exceeds the CA validity.")
}

// Positive test for whole workflow.
// Make sure all the ENV are exported.
func TestAWSProvider(t *testing.T) {
	rootCAIdentifier := os.Getenv("ROOT_CA_ARN")
	templateArn := os.Getenv("TEMPLATE_ARN")

	cmd := exec.Command("../getistio", "gen-ca", "--provider=aws",
		fmt.Sprintf("--signing-ca=%s", rootCAIdentifier),
		fmt.Sprintf("--template-arn=%s", templateArn),
		"--validity-days=100",
		"--secret-file-path=/tmp/file2.yaml",
	)
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	cmd.Run()
	actual := buf.String()
	assert.Contains(t, actual, "apiVersion: v1\nkind: Secret\nmetadata:\n  name: cacerts\n  namespace: istio-system\ntype: Opaque\ndata:\n  ca-cert.pem: -----BEGIN CERTIFICATE-----\n")
}
