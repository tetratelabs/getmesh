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
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	kiali_business "github.com/kiali/kiali/business"
	kiali_config "github.com/kiali/kiali/config"
	kiali_kubernetes "github.com/kiali/kiali/kubernetes"
	kiali_log "github.com/kiali/kiali/log"
	kiali_models "github.com/kiali/kiali/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/tetratelabs/getistio/src/util"
)

func init() {
	kiali_log.InitializeLogger().Output(ioutil.Discard)
	kiali_config.Set(kiali_config.NewConfig())
}

// KialiConfigValidations returns Istio Validation Summary for the selected namespace and service requested.
func (cv *ConfigValidator) kialiConfigValidations() ([]configValidationResult, error) {
	config, err := util.GetK8sConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconf")
	}

	raw, err := kiali_kubernetes.NewClientFromConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate kiali client")
	}

	wrapper, err := newKialiClientWrapper(raw, cv.files, cv.namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create wrapper of Kiali client: %w", err)
	}

	kialiBackend := kiali_business.NewWithBackends(wrapper, nil, nil)
	var ret []configValidationResult
	if cv.allNamespaces() {
		ret, err = kialiAllNamespaceConfigValidation(cv.kubeCli, kialiBackend.Validations)
	} else {
		ret, err = kialiSingleNamespaceConfigValidation(kialiBackend.Validations, cv.namespace)
	}

	if err != nil {
		return nil, err
	}

	for i, r := range ret {
		// check if there's file reference for local file validations, and if found, append the information to messages
		// just like Istioctl analyze does for local files
		key := kialiObjectKey(r.namespace, r.name, r.resourceType)
		if f, ok := wrapper.localIstioObjectFilesRef[key]; ok {
			ret[i].message = fmt.Sprintf("[%s] %s", f, r.message)
		}
	}
	return ret, nil

}

func kialiAllNamespaceConfigValidation(kubeCli kubernetes.Interface,
	kialiIstioValidationService kiali_business.IstioValidationsService) ([]configValidationResult, error) {
	nsList, err := kubeCli.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var (
		vs        = kiali_models.IstioValidations{}
		errorList []error
	)

	for _, ns := range nsList.Items {
		validations, err := kialiIstioValidationService.GetValidations(ns.Name, "")
		if err != nil && !strings.Contains(err.Error(), "excluded for Kiali") {
			errorList = append(errorList, err)
			continue
		}

		vs.MergeValidations(validations)
	}

	if len(errorList) != 0 {
		return nil, util.HandleMultipleErrors(errorList)
	}

	return convertKialiModelToResult(vs), nil
}

func kialiSingleNamespaceConfigValidation(kialiIstioValidationService kiali_business.IstioValidationsService,
	namespace string) ([]configValidationResult, error) {
	vs, err := kialiIstioValidationService.GetValidations(namespace, "")

	if err != nil {
		return nil, fmt.Errorf("failed to get kiali validation: %w", err)
	}
	return convertKialiModelToResult(vs), nil
}

func convertKialiModelToResult(validation kiali_models.IstioValidations) []configValidationResult {
	var res []configValidationResult
	for key, v := range validation {
		for _, c := range v.Checks {
			ps := strings.SplitN(c.Message, " ", 2)
			if len(ps) != 2 {
				continue
			}
			res = append(res, configValidationResult{
				name:         key.Name,
				namespace:    key.Namespace,
				resourceType: v.ObjectType,
				severity:     convertKialiSeverity(c.Severity),
				message:      ps[1],
				errorCode:    ps[0],
			})
		}
	}
	return res
}

func convertKialiSeverity(in kiali_models.SeverityLevel) Severity {
	switch string(in) {
	case "unknown":
		return SeverityLevelInfo
	case "warning", "warn":
		return SeverityLevelWarn
	case "error":
		return SeverityLevelError
	default:
		return SeverityLevelUnknown
	}
}
