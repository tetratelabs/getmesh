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

package manifest

import (
	"github.com/tetratelabs/getmesh/api"
	"github.com/tetratelabs/getmesh/src/istioctl"
	"github.com/tetratelabs/getmesh/src/util"
	"github.com/tetratelabs/getmesh/src/util/logger"
)

func securityPatchChecker(m *api.Manifest) error {
	hd, err := util.GetmeshHomeDir()
	if err != nil {
		return err
	}
	return securityPatchCheckerImpl(hd, m)
}

func securityPatchCheckerImpl(homedir string, m *api.Manifest) error {
	vs, err := istioctl.GetFetchedVersions(homedir)
	if err != nil {
		return err
	}

	// construct {version's group -> highest lowest patch version in the group} map in the locally installed versions
	locals, err := constructLatestVersionsMap(vs)
	if err != nil {
		return err
	}

	for g, local := range locals {
		target, includeSecurityPatch, err := findSecurityPatchUpgrade(local, g, m.IstioDistributions)
		if err != nil {
			return err
		}

		if target == nil {
			logger.Warnf("The locally installed minor version %s is no longer supported by getmesh. "+
				"We recommend you use the higher minor versions in \"getmesh list\" or remove with \"getmesh prune\"\n", g)
			continue
		}

		greater, err := target.GreaterThan(local)
		if err != nil {
			return err
		} else if greater && includeSecurityPatch {
			t := target.ToString()
			logger.Warnf("The locally installed minor version %s has a latest version %s including security patches. "+
				"We strongly recommend you to download %s by \"getmesh fetch\".\n", g, t, t)
		}
	}

	return nil
}

func constructLatestVersionsMap(in []*api.IstioDistribution) (map[string]*api.IstioDistribution, error) {
	ret := map[string]*api.IstioDistribution{}
	for _, v := range in {
		vg, err := v.Group()
		if err != nil {
			return nil, err
		}
		prev, ok := ret[vg]
		if !ok {
			ret[vg] = v
			continue
		}

		ok, err = v.GreaterThan(prev)
		if err != nil {
			return nil, err
		} else if ok {
			ret[vg] = v
		}
	}
	return ret, nil
}

// walk thorough distributions in the manifest higher than "base" distribution in its minor version.
func findSecurityPatchUpgrade(base *api.IstioDistribution, group string, remotes []*api.IstioDistribution) (
	target *api.IstioDistribution, includeSecurityPatch bool, error error) {

	for _, r := range remotes {
		rg, err := r.Group()
		if err != nil {
			return nil, false, err
		}

		if rg != group {
			continue
		}

		greater, err := r.GreaterThan(base)
		if err != nil {
			return nil, false, err
		} else if !greater && !r.Equal(base) {
			continue
		}

		if r.IsSecurityPatch {
			includeSecurityPatch = true
		}

		if target == nil {
			target = r
			continue
		}

		greater, err = r.GreaterThan(target)
		if err != nil {
			return nil, false, err
		}

		if greater {
			target = r
		}
	}
	return
}
