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
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestParseFilesAsKialiIstioObjects(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	_, err = f.WriteString(`
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: a
  namespace: default
spec:
  hosts:
    - reviews
  http:
    - route:
        - destination:
            host: reviews
            subset: v1
          weight: 10
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: b
  namespace: default
spec:
  hosts:
    - reviews
  http:
    - route:
        - destination:
            host: reviews
            subset: v1
          weight: 10
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: c
  namespace: default
spec:
  hosts:
    - reviews
  http:
    - route:
        - destination:
            host: reviews
            subset: v1
          weight: 10
---
a: a
---
b: b
---
`)
	require.NoError(t, err)

	g, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	_, err = g.WriteString(`
apiVersion: v1
kind: Service
metadata:
  name: httpbin
  namespace: healthy
  labels:
    app: httpbin
  annotations:
    # no such Istio annotation
    networking.istio.io/non-exist: bar
spec:
  ports:
    - name: http
      port: 8000
      targetPort: 80
  selector:
    app: httpbin
`) // ignored
	require.NoError(t, err)

	c := kialiClientWrapper{}

	require.NoError(t, c.parseFilesAsKialiIstioObjects([]string{f.Name(), g.Name()}, ""))

	for _, exp := range []string{
		"a", "b", "c",
	} {

		var found bool
		for _, obj := range c.localIstioObjects[kialiObjectListKey("default", "VirtualService")] {
			if obj.GetObjectMeta().Name == exp {
				found = true
				break
			}
		}
		assert.True(t, found)
		assert.Contains(t, c.localIstioObjectFilesRef, kialiObjectKey("default", exp, "VirtualService"))
	}
}
