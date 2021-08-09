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

package cmd

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getmesh/internal/manifest"
)

func Test_pruneCheckFlags(t *testing.T) {
	for i, c := range []struct {
		version, flavor string
		flavorVersion   int
		exp             *manifest.IstioDistribution
		expErr          bool
	}{
		{version: "", flavor: "", flavorVersion: -1, expErr: false, exp: nil},
		{version: "1.7.0", flavor: "", flavorVersion: -1, expErr: true},
		{version: "", flavor: "tetrate", flavorVersion: -1, expErr: true},
		{version: "", flavor: "", flavorVersion: 1, expErr: true},
		{version: "1.7.0", flavor: "tetrate", flavorVersion: -1, expErr: true},
		{version: "", flavor: "tetrate", flavorVersion: 1, expErr: true},
		{version: "1.7.0", flavor: "tetrate", flavorVersion: 1, expErr: false, exp: &manifest.IstioDistribution{
			Version:       "1.7.0",
			Flavor:        "tetrate",
			FlavorVersion: 1,
		}},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual, err := pruneCheckFlags(c.version, c.flavor, c.flavorVersion)
			if c.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, c.exp, actual)
		})
	}
}
