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
//
// Some of code here has been borrowed from Kiali:
//   https://github.com/kiali/kiali/blob/6ec37a53ddbcc88ee55959aeeac7114b33f893aa/kubernetes/istio.go#L113-L152
// It is under Apache License 2.0.

package configvalidator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	kiali_kubernetes "github.com/kiali/kiali/kubernetes"
	"golang.org/x/sync/singleflight"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/tetratelabs/getmesh/src/util"
	"github.com/tetratelabs/getmesh/src/util/logger"
)

// There are two reasons for having kialiClientWrapper:
//
// 1) kialiClientWrapper overrides the original getmeshObjects method. This is necessary because
// 	Kiali somehow re-creates Kubeconfig from the original one which results in the `system:anonoymous` user error.
// 	See https://github.com/kiali/kiali/blob/6ec37a53ddbcc88ee55959aeeac7114b33f893aa/kubernetes/client.go#L286-L315.
//
// 2) in order to do the local yaml validation with Kiali
type kialiClientWrapper struct {
	*kiali_kubernetes.K8SClient
	networkingAPI, securityAPI *rest.RESTClient
	k8s                        *kubernetes.Clientset
	sf                         singleflight.Group

	// read only
	localIstioObjects        map[string]map[string]kiali_kubernetes.IstioObject
	localIstioObjectFilesRef map[string]string
}

func init() {
	istioTypes = make(map[string]struct{}, len(networkingTypes)+len(securityTypes))
	for _, t := range networkingTypes {
		istioTypes[t.objectKind] = struct{}{}
	}
	for _, t := range securityTypes {
		istioTypes[t.objectKind] = struct{}{}
	}
}

func (c *kialiClientWrapper) parseFilesAsKialiIstioObjects(files []string, namespace string) error {
	objs := map[string]map[string]kiali_kubernetes.IstioObject{}
	refs := map[string]string{}
	for _, f := range files {
		raw, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}

		r := yaml.NewDecoder(bytes.NewReader(raw))
		for {

			var obj kiali_kubernetes.GenericIstioObject
			if err := r.Decode(&obj); err != nil && err != io.EOF {
				return fmt.Errorf("failed to unmarshal %s: %v", f, err)
			} else if err == io.EOF {
				break
			}

			kind := obj.GetObjectKind().GroupVersionKind().Kind
			if _, ok := istioTypes[kind]; !ok {
				continue
			}

			if obj.Namespace == "" {
				if namespace != "" {
					obj.Namespace = namespace
				} else {
					obj.Namespace = "default"
				}
				logger.Infof("The object %s in %s does not have namespace so we assume it's applied to \"%s\" namespace \n",
					obj.Name, f, obj.Namespace)
			}
			key := kialiObjectListKey(obj.Namespace, kind)

			m, ok := objs[key]
			if !ok {
				m = map[string]kiali_kubernetes.IstioObject{}
			}

			objKey := kialiObjectKey(obj.Namespace, obj.Name, kind)
			m[objKey] = &obj
			refs[objKey] = f
			objs[key] = m
		}
	}

	c.localIstioObjects = objs
	c.localIstioObjectFilesRef = refs
	return nil
}

var (
	networkingTypes = []struct {
		objectKind     string
		collectionKind string
	}{
		{
			objectKind:     kiali_kubernetes.GatewayType,
			collectionKind: kiali_kubernetes.GatewayTypeList,
		},
		{
			objectKind:     kiali_kubernetes.VirtualServiceType,
			collectionKind: kiali_kubernetes.VirtualServiceTypeList,
		},
		{
			objectKind:     kiali_kubernetes.DestinationRuleType,
			collectionKind: kiali_kubernetes.DestinationRuleTypeList,
		},
		{
			objectKind:     kiali_kubernetes.ServiceEntryType,
			collectionKind: kiali_kubernetes.ServiceentryTypeList,
		},
		{
			objectKind:     kiali_kubernetes.SidecarType,
			collectionKind: kiali_kubernetes.SidecarTypeList,
		},
		{
			objectKind:     kiali_kubernetes.WorkloadEntryType,
			collectionKind: kiali_kubernetes.WorkloadEntryTypeList,
		},
		{
			objectKind:     kiali_kubernetes.EnvoyFilterType,
			collectionKind: kiali_kubernetes.EnvoyFilterTypeList,
		},
	}
	securityTypes = []struct {
		objectKind     string
		collectionKind string
	}{
		{
			objectKind:     kiali_kubernetes.PeerAuthenticationsType,
			collectionKind: kiali_kubernetes.PeerAuthenticationsTypeList,
		},
		{
			objectKind:     kiali_kubernetes.AuthorizationPoliciesType,
			collectionKind: kiali_kubernetes.AuthorizationPoliciesTypeList,
		},
		{
			objectKind:     kiali_kubernetes.RequestAuthenticationsType,
			collectionKind: kiali_kubernetes.RequestAuthenticationsTypeList,
		},
	}

	istioTypes map[string]struct{}
)

func newKialiClientWrapper(base *kiali_kubernetes.K8SClient, files []string, namespace string) (*kialiClientWrapper, error) {
	sc := runtime.NewScheme()
	for _, nt := range networkingTypes {
		sc.AddKnownTypeWithName(kiali_kubernetes.NetworkingGroupVersion.WithKind(nt.objectKind), &kiali_kubernetes.GenericIstioObject{})
		sc.AddKnownTypeWithName(kiali_kubernetes.NetworkingGroupVersion.WithKind(nt.collectionKind), &kiali_kubernetes.GenericIstioObjectList{})
	}
	for _, rt := range securityTypes {
		sc.AddKnownTypeWithName(kiali_kubernetes.SecurityGroupVersion.WithKind(rt.objectKind), &kiali_kubernetes.GenericIstioObject{})
		sc.AddKnownTypeWithName(kiali_kubernetes.SecurityGroupVersion.WithKind(rt.collectionKind), &kiali_kubernetes.GenericIstioObjectList{})
	}
	metav1.AddToGroupVersion(sc, kiali_kubernetes.NetworkingGroupVersion)
	metav1.AddToGroupVersion(sc, kiali_kubernetes.SecurityGroupVersion)

	k8s, err := util.GetK8sClient()
	if err != nil {
		return nil, err
	}

	ret := &kialiClientWrapper{K8SClient: base, k8s: k8s}
	if err := ret.parseFilesAsKialiIstioObjects(files, namespace); err != nil {
		return nil, fmt.Errorf("failed to prase local yaml files: %s", err)
	}

	// create networking api client
	config, err := util.GetK8sConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconf")
	}

	config.ContentConfig = rest.ContentConfig{
		GroupVersion:         &kiali_kubernetes.NetworkingGroupVersion,
		NegotiatedSerializer: serializer.WithoutConversionCodecFactory{CodecFactory: serializer.NewCodecFactory(sc)},
		ContentType:          runtime.ContentTypeJSON,
	}
	config.APIPath = "/apis"

	if ret.networkingAPI, err = rest.RESTClientFor(config); err != nil {
		return nil, fmt.Errorf("failed to create client for istio networking api: %w", err)
	}

	// create security api client
	config, err = util.GetK8sConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconf")
	}

	config.APIPath = "/apis"

	config.ContentConfig = rest.ContentConfig{
		GroupVersion:         &kiali_kubernetes.SecurityGroupVersion,
		NegotiatedSerializer: serializer.WithoutConversionCodecFactory{CodecFactory: serializer.NewCodecFactory(sc)},
		ContentType:          runtime.ContentTypeJSON,
	}

	if ret.securityAPI, err = rest.RESTClientFor(config); err != nil {
		return nil, fmt.Errorf("failed to create client for istio security api: %w", err)
	}
	return ret, nil
}

func (c *kialiClientWrapper) getAPIClientVersion(apiGroup string) (*rest.RESTClient, string) {
	if apiGroup == kiali_kubernetes.NetworkingGroupVersion.Group {
		return c.networkingAPI, kiali_kubernetes.ApiNetworkingVersion
	} else if apiGroup == kiali_kubernetes.SecurityGroupVersion.Group {
		return c.securityAPI, kiali_kubernetes.ApiSecurityVersion
	}
	return nil, ""
}

// override the original "dead" getmeshObjects
func (c *kialiClientWrapper) getmeshObjects(namespace, resourceType, labelSelector string) ([]kiali_kubernetes.IstioObject, error) {
	kind := kiali_kubernetes.PluralType[resourceType]
	key := kialiObjectListKey(namespace, kind)
	raw, err, _ := c.sf.Do(key, func() (interface{}, error) {
		locals, ok := c.localIstioObjects[key]
		ret := make([]kiali_kubernetes.IstioObject, 0, len(locals))
		if ok {
			for _, l := range locals {
				ret = append(ret, l)
			}
		}

		var apiClient *rest.RESTClient
		var apiGroup string
		if apiGroup, ok = kiali_kubernetes.ResourceTypesToAPI[resourceType]; ok {
			apiClient, _ = c.getAPIClientVersion(apiGroup)
		} else {
			return nil, fmt.Errorf("%s not found in ResourcesTypeToAPI", resourceType)
		}

		var result runtime.Object
		var err error
		result, err = apiClient.Get().Namespace(namespace).Resource(resourceType).
			Param("labelSelector", labelSelector).Do(context.Background()).Get()
		if err != nil {
			if ss, ok := err.(errors.APIStatus); ok && ss.Status().Code == http.StatusNotFound {
				logger.Infof("Resource %s not found in the cluster."+
					" Maybe Istio has not been installed in your cluster."+
					" Please make sure Istio is installed before execute \"getmesh config-validate\"\n", kind)
				os.Exit(1)
			}
			return nil, err
		}

		istioList, ok := result.(*kiali_kubernetes.GenericIstioObjectList)
		if !ok {
			return nil, fmt.Errorf("%s/%s doesn't return a list", namespace, resourceType)
		}

		for _, item := range istioList.GetItems() {
			m := item.GetObjectMeta()
			if _, ok := locals[kialiObjectKey(m.Namespace, m.Name, kind)]; ok {
				// remote objects are overridden by local ones
				continue
			}
			ret = append(ret, item)
		}
		return ret, nil
	})

	if err != nil {
		return nil, err
	}

	return raw.([]kiali_kubernetes.IstioObject), nil
}

func kialiObjectListKey(namespace, kind string) string {
	return fmt.Sprintf("%s.%s", namespace, strings.ToLower(kind))
}

func kialiObjectKey(namespace, name, kind string) string {
	return fmt.Sprintf("%s.%s.%s", namespace, name, strings.ToLower(kind))
}
