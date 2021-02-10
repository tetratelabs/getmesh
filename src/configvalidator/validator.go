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
	"errors"
	"fmt"

	"k8s.io/client-go/kubernetes"

	"github.com/tetratelabs/getistio/src/util"
	"github.com/tetratelabs/getistio/src/util/logger"
)

var ErrConfigIssuesFound = errors.New("getistio config validation exit with istio config issues")

// ConfigValidator is general structure for validating istio config
// preset in current live-cluster
type ConfigValidator struct {
	kubeCli         kubernetes.Interface
	namespace       string
	getistioHomedir string
	outputThreshold Severity
	files           []string
}

// InitConfigValidator initialize the ConfigValidator struct.
func New(homedir, namespace, outputThreshold string, files []string) (*ConfigValidator, error) {
	kubeCli, err := util.GetK8sClient()
	if err != nil {
		return nil, fmt.Errorf("error getting k8s client: %w", err)
	}

	files, err = extractYamlFilePaths(files)
	if err != nil {
		return nil, fmt.Errorf("error walking thourgh paths: %v", err)
	}

	sv, ok := strToSeverityLevel[outputThreshold]
	if !ok {
		return nil, fmt.Errorf("invalid output-threshold %s", outputThreshold)
	}

	return &ConfigValidator{
		kubeCli:         kubeCli,
		namespace:       namespace,
		getistioHomedir: homedir,
		outputThreshold: sv,
		files:           files,
	}, nil
}

func (cv *ConfigValidator) Validate() error {
	msg := "Running the config validator"
	if !cv.allNamespaces() {
		msg += " for namespace=" + cv.namespace
	}

	logger.Infof(msg + ". This may take some time...\n\n")

	kvs, err := cv.kialiConfigValidations()
	if err != nil {
		return fmt.Errorf("error kiali validation: %w", err)
	}

	ivs, err := cv.istioAnalyseValidations()
	if err != nil {
		return fmt.Errorf("error istioctl validation: %w", err)
	}

	results := cv.filterResults(append(ivs, kvs...))
	if len(results) == 0 {
		logger.Infof("Your Istio configurations are healthy. Configuration issues not found.\n")
		return nil
	}

	if cv.allNamespaces() {
		printResults(results)
	} else {
		printResultsWithoutNamespace(results)
	}

	logger.Infof(`
The error codes of the found issues are prefixed by 'IST' or 'KIA'. For the detailed explanation, please refer to
- https://istio.io/latest/docs/reference/config/analysis/ for 'IST' error codes
- https://kiali.io/documentation/latest/validations/ for 'KIA' error codes
`)
	return ErrConfigIssuesFound
}

func (cv *ConfigValidator) filterResults(in []configValidationResult) []configValidationResult {
	out := make([]configValidationResult, 0, len(in))
	for _, r := range in {
		if (cv.allNamespaces() || (r.namespace == cv.namespace)) && (r.severity.level <= cv.outputThreshold.level) {
			out = append(out, r)
		}
	}
	return out
}

func (cv *ConfigValidator) allNamespaces() bool {
	return cv.namespace == ""
}
