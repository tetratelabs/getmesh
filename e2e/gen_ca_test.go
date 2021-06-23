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

package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckGenCAConfigFileNotExist(t *testing.T) {
	cmd := exec.Command("./getmesh", "gen-ca", "--config-file=/tmp/filenotexist.yaml")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
	actual := buf.String()
	require.Contains(t, actual, `unable to open config file path: open /tmp/filenotexist.yaml: no such file or directory`)
}

func TestCheckGenCAConfigProviderUnavailable(t *testing.T) {
	cmd := exec.Command("./getmesh", "gen-ca", "--provider=unavailable")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
	actual := buf.String()
	require.Contains(t, actual, "`unavailable` provider yet to be implement")
}

func TestCheckGenCAConfigSigningCANotProvided1(t *testing.T) {
	cmd := exec.Command("./getmesh", "gen-ca", "--provider=aws")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
	actual := buf.String()
	require.Contains(t, actual, "found empty AWS Signing CA ARN")
}

func TestCheckGenCAConfigSigningCANotProvided2(t *testing.T) {
	cmd := exec.Command("./getmesh", "gen-ca", "--provider=aws")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
	actual := buf.String()
	require.Contains(t, actual, "found empty AWS Signing CA ARN")
}

func TestCheckGenCAConfigWrongRegionProvided(t *testing.T) {
	cmd := exec.Command("./getmesh", "gen-ca", "--provider=aws", "--signing-ca=testing", "--secret-file-path=/tmp/temp.yaml")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
	actual := buf.String()
	require.Contains(t, actual, "unable to issue CA, due to error: unable to initialize AWS Client: unable to get AWS Region: unable to find region for AWS Signing CA ARN")
}

func TestCheckGenCAConfigWrongInfoProvided(t *testing.T) {
	cmd := exec.Command("./getmesh", "gen-ca", "--provider=aws", "--signing-ca=arn:aws:acm-pca:us-west-2:123456789:certificate-authority/fake", "--secret-file-path=/tmp/temp.yaml")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
	actual := buf.String()
	require.Contains(t, actual, "unable to issue CA, due to error: unable to get CACertificateRequest: NoCredentialProviders")
}

func TestCheckGenCAConfigWrongFlagsProvided(t *testing.T) {
	cmd := exec.Command("./getmesh", "gen-ca", "--testing=test")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
	actual := buf.String()
	require.Contains(t, actual, "unknown flag: --testing\n")
}
