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

package util

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var KubeConfig string

func GetKubeConfigLocation() string {
	kubeconfig := KubeConfig
	if kubeconfig != "" {
		return kubeconfig
	}
	kubeconfig = os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	}
	return kubeconfig
}

func GetK8sConfig() (*rest.Config, error) {
	kubeconfig := GetKubeConfigLocation()
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building config from kubeconfig located in %s: %w", kubeconfig, err)
	}
	return config, nil
}

func GetK8sClient() (*kubernetes.Clientset, error) {
	config, err := GetK8sConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	kubeCli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate k8s client: %w", err)
	}

	return kubeCli, nil
}
