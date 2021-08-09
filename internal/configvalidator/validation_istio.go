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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/tetratelabs/getmesh/internal/istioctl"
)

func (cv *ConfigValidator) istioAnalyseValidations() ([]configValidationResult, error) {
	var analyzeCommandArgs = []string{"analyze", "--color=false"}

	if cv.allNamespaces() {
		analyzeCommandArgs = append(analyzeCommandArgs, "--all-namespaces")
	} else {
		analyzeCommandArgs = append(analyzeCommandArgs, fmt.Sprintf("--namespace=%s", cv.namespace))
	}

	if len(cv.files) > 0 {
		analyzeCommandArgs = append(analyzeCommandArgs, "-R")
		analyzeCommandArgs = append(analyzeCommandArgs, cv.files...)
	}

	out := new(bytes.Buffer)
	stde := new(bytes.Buffer)
	err := istioctl.ExecWithWriters(cv.getmeshHomedir, analyzeCommandArgs, out, stde)
	if err != nil && !strings.Contains(stde.String(), "Error: Analyzers found issues when") {
		return nil, errors.New(stde.String())
	}

	return parseIstioctlAnalyzeResult(out), nil
}

func parseIstioctlAnalyzeResult(r io.Reader) []configValidationResult {
	var res []configValidationResult

	// Parse the message line by line according to the specification:
	// TODO: that this may be fragile once the format changes.
	// https://istio.io/latest/docs/reference/config/analysis/message-format/
	b := bufio.NewScanner(r)
	for b.Scan() {
		l := b.Text()

		tokens := strings.SplitN(l, " ", 3)
		if len(tokens) != 3 {
			// TODO: debug logging
			continue
		}
		level, code := tokens[0], tokens[1]

		var affectedResource, msg string
		right := strings.TrimLeft(tokens[2], "(")
		if n := strings.Index(right, ") "); n < 0 {
			// TODO: debug logging
			continue
		} else {
			affectedResource = right[:n]
			msg = right[n+2:]
		}

		tokens = strings.SplitN(affectedResource, " ", 2)
		if len(tokens) != 2 {
			// TODO: debug logging
			continue
		}
		resourceKind, resourceID := tokens[0], tokens[1]

		tokens = strings.SplitN(resourceID, ".", 2)
		var name, namespace string
		if len(tokens) != 2 && resourceKind != "Namespace" {
			// TODO: debug logging
			continue
		} else if len(tokens) != 2 {
			// if the target resource kind is namespace, then <resource-namespace> is omitted
			name, namespace = resourceID, resourceID
		} else {
			name, namespace = tokens[0], tokens[1]
		}

		// if name contains the white space, this is the case of local file without namespace begin specified
		if tk := strings.SplitN(name, " ", 2); len(tk) > 1 {
			msg = fmt.Sprintf("[%s] %s", tk[1]+"."+namespace, msg)
			name = tk[0]
			namespace = ""
		}

		// if namespace contains spaces, then it is a local file validation and contains the file path.
		if tk := strings.SplitN(namespace, " ", 2); len(tk) > 1 {
			namespace = tk[0]
			msg = fmt.Sprintf("[%s] %s", tk[1], msg)
		}

		res = append(res, configValidationResult{
			name:         name,
			namespace:    namespace,
			resourceType: resourceKind,
			severity:     convertIstioLevel(level),
			message:      msg,
			errorCode:    code[1 : len(code)-1],
		})
	}
	return res
}

func convertIstioLevel(in string) Severity {
	switch in {
	case "Info":
		return SeverityLevelInfo
	case "Warn", "Warning":
		return SeverityLevelWarn
	case "Error":
		return SeverityLevelError
	default:
		return SeverityLevelUnknown
	}
}
