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

package configvalidator

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIstioctlAnalyzeResult(t *testing.T) {
	for _, c := range []struct {
		in  string
		exp configValidationResult
	}{
		{
			// remote validation
			in: `Error [IST0101] (VirtualService details.bookinfo) Referenced host+subset in destinationrule not found: "details+v1"`,
			exp: configValidationResult{
				name:         "details",
				namespace:    "bookinfo",
				errorCode:    "IST0101",
				resourceType: "VirtualService",
				message:      "Referenced host+subset in destinationrule not found: \"details+v1\"",
				severity:     SeverityLevelError,
			},
		},
		{
			// local file validation with namespace case
			in: `Error [IST0106] (VirtualService ratings-bogus-weight-default.healthy e2e/testdata/config-validate-local.yaml:1) Schema validation error: total destination weight 1887 != 100`,
			exp: configValidationResult{
				name:         "ratings-bogus-weight-default",
				namespace:    "healthy",
				errorCode:    "IST0106",
				resourceType: "VirtualService",
				message:      "[e2e/testdata/config-validate-local.yaml:1] Schema validation error: total destination weight 1887 != 100",
				severity:     SeverityLevelError,
			},
		},
		{
			// local file validation without namespace case
			in: `Error [IST0106] (VirtualService ratings-bogus-weight-default e2e/testdata/config-validate-local.yaml:1) Schema validation error: total destination weight 1887 != 100`,
			exp: configValidationResult{
				name:         "ratings-bogus-weight-default",
				namespace:    "",
				errorCode:    "IST0106",
				resourceType: "VirtualService",
				message:      "[e2e/testdata/config-validate-local.yaml:1] Schema validation error: total destination weight 1887 != 100",
				severity:     SeverityLevelError,
			},
		},
	} {
		assert.Equal(t, []configValidationResult{c.exp},
			parseIstioctlAnalyzeResult(bytes.NewBufferString(c.in)), c.in)
	}
}
