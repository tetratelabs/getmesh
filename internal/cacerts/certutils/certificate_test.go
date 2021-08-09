// Copyright 2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package certutils

import (
	"crypto/x509"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateCertificateSigningRequest(t *testing.T) {
	csrBytesEncoded, subordinateCaKey, err := CreateCSR(x509.CertificateRequest{}, 2048)
	require.NoError(t, err)
	require.Equal(t, strings.HasPrefix(string(csrBytesEncoded), "-----BEGIN CERTIFICATE REQUEST-----"), true)
	require.Equal(t, strings.HasPrefix(subordinateCaKey, "-----BEGIN PRIVATE KEY-----"), true)
}
