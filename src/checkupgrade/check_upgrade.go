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

package checkupgrade

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	// https://github.com/istio/pkg/blob/4f521de9c8caa220ebc9e7f57da2726dff2788fc/version/cobra.go
	// TODO: Though this package is stable and it's been over a year since it changed last (as of 2020/11/19),
	// 	using this may be fragile due to the I/F change in the future
	istioversion "istio.io/pkg/version"

	"github.com/tetratelabs/getmesh/api"
	"github.com/tetratelabs/getmesh/src/util/logger"
)

var ErrIssueFound = errors.New("version issue found")

func IstioVersion(iv istioversion.Version, manifest *api.Manifest) error {
	logger.Infof("[Summary of your Istio mesh]\n")
	printSummary(iv)
	logger.Infof("[GetMesh Check]\n")
	return printgetmeshCheck(iv, manifest)
}

func printSummary(iv istioversion.Version) {
	msg := fmt.Sprintf("active istioctl version: %s\n", iv.ClientVersion.Version)

	if dv := iv.DataPlaneVersion; dv != nil && len(*dv) > 0 {
		versions := make(map[string]int, len(*dv))
		for _, info := range *dv {
			versions[info.IstioVersion]++
		}
		counts := make([]string, 0, len(versions))
		for ver, num := range versions {
			counts = append(counts, fmt.Sprintf("%s (%d proxies)", ver, num))
		}
		sort.Strings(counts) // to make deterministic
		msg += fmt.Sprintf("data plane version: %s\n", strings.Join(counts, ", "))
	}

	if mv := iv.MeshVersion; mv != nil && len(*mv) > 0 {
		vm := make(map[string]struct{}, len(*mv))
		vs := make([]string, 0, len(*mv))
		for _, m := range *mv {
			if _, ok := vm[m.Info.Version]; !ok {
				vm[m.Info.Version] = struct{}{}
				vs = append(vs, m.Info.Version)
			}
		}

		sort.Strings(vs) // to make deterministic
		msg += fmt.Sprintf("control plane version: %s\n", strings.Join(vs, ", "))
	}

	logger.Infof("%s\n", msg)
}

func printgetmeshCheck(iv istioversion.Version, manifest *api.Manifest) error {
	var multiple bool
	dpVersions, err := getDataPlaneVersions(iv.DataPlaneVersion)
	if err != nil {
		return fmt.Errorf("collecting data plane versions: %v", err)
	}

	cpVersions, err := getControlPlaneVersions(iv.MeshVersion)
	if err != nil {
		return fmt.Errorf("collecting control plane versions: %v", err)
	}

	if len(dpVersions) > 1 {
		multiple = true
		logger.Infof(getMultipleMinorVersionRunningMsg("data plane", dpVersions))
	}

	if len(cpVersions) > 1 {
		multiple = true
		logger.Infof(getMultipleMinorVersionRunningMsg("control plane", cpVersions))
	}

	if len(cpVersions) == 0 && len(dpVersions) == 0 {
		logger.Infof("nothing to check.\n")
		return nil
	}

	if !multiple &&
		(len(cpVersions) == len(dpVersions) && len(cpVersions) == 1) {

		var cpV, dpV string
		for v := range cpVersions {
			cpV = v
		}
		for v := range dpVersions {
			dpV = v
		}
		if cpV != dpV {
			logger.Infof("- Your data plane running in the minor version %s but control plane in %s\n", dpV, cpV)
		}
	}

	versionToLowestPatches := dpVersions
	for k, next := range cpVersions {
		if prev, ok := versionToLowestPatches[k]; !ok {
			versionToLowestPatches[k] = next
		} else if ok, _ = prev.GreaterThan(next); ok {
			versionToLowestPatches[k] = next
		}
	}

	var okCount int
	for group, v := range versionToLowestPatches {
		msg, ok, err := getLatestPatchInManifestMsg(v, manifest)
		if err != nil {
			return fmt.Errorf("checking the latest patch for %s: %v", group, err)
		}
		if ok {
			okCount++
		}
		logger.Infof(msg)
	}

	if okCount == len(versionToLowestPatches) {
		return nil
	}
	return ErrIssueFound
}

func getLatestPatchInManifestMsg(target *api.IstioDistribution, manifest *api.Manifest) (string, bool, error) {
	tg, err := target.Group()
	if err != nil {
		return "", false, err
	}

	foundLatest, includeSecurityPatch, err := api.GetLatestDistribution(target, manifest)
	if err != nil {
		return "", false, err
	}

	if foundLatest == nil {
		return fmt.Sprintf("- The minor version %s is no longer supported by getmesh. "+
			"We recommend you use the higher minor versions in \"getmesh list\"\n", tg), false, nil
	}

	if foundLatest.Equal(target) {
		return fmt.Sprintf("- %s is the latest version in %s\n", target.ToString(), tg), true, nil
	}

	msg := fmt.Sprintf("- There is the available patch for the minor version %s", tg)
	if includeSecurityPatch {
		msg += fmt.Sprintf(" which includes **security upgrades**. We strongly recommend upgrading all %s versions -> %s\n", tg, foundLatest.ToString())
	} else {
		msg += fmt.Sprintf(". We recommend upgrading all %s versions -> %s\n", tg, foundLatest.ToString())

	}
	return msg, false, nil
}

func getMultipleMinorVersionRunningMsg(t string, mvs map[string]*api.IstioDistribution) string {
	const template = "- Your %s running in multiple minor versions: %s\n"
	vs := make([]string, 0, len(mvs))
	for v := range mvs {
		vs = append(vs, v)
	}
	sort.Strings(vs)
	return fmt.Sprintf(template, t, strings.Join(vs, ", "))
}

// construct {version's group -> lowest patch version} map from control plane versions
func getControlPlaneVersions(info *istioversion.MeshInfo) (map[string]*api.IstioDistribution, error) {
	if info == nil {
		return map[string]*api.IstioDistribution{}, nil
	}

	in := make([]*api.IstioDistribution, len(*info))
	for i, raw := range *info {
		v, err := api.IstioDistributionFromString(raw.Info.Version)
		if err != nil {
			return nil, fmt.Errorf("error parsing control's version %s: %v", raw.Info.Version, err)
		}
		in[i] = v
	}
	return findLowestPatchVersionsInGroup(in)
}

// construct {version's group -> lowest patch version} map from data plane versions
func getDataPlaneVersions(info *[]istioversion.ProxyInfo) (map[string]*api.IstioDistribution, error) {
	if info == nil {
		return map[string]*api.IstioDistribution{}, nil
	}

	in := make([]*api.IstioDistribution, len(*info))
	for i, raw := range *info {
		v, err := api.IstioDistributionFromString(raw.IstioVersion)
		if err != nil {
			return nil, fmt.Errorf("error parsing dataplane's version %s: %v", raw.IstioVersion, err)
		}
		in[i] = v
	}

	return findLowestPatchVersionsInGroup(in)
}

func findLowestPatchVersionsInGroup(in []*api.IstioDistribution) (map[string]*api.IstioDistribution, error) {
	ret := make(map[string]*api.IstioDistribution, len(in))
	for _, d := range in {
		if d.IsUpstream() {
			logger.Warnf("the upstream istio distributions are not supported by check-upgrade command: %s. "+
				"Please install distributions with tetrate flavor listed in `getmesh list` command\n", d.Version)
			continue
		}

		vg, err := d.Group()
		if err != nil {
			return nil, fmt.Errorf("error parsing version %s: %v", d.ToString(), err)
		}
		prev, ok := ret[vg]
		if !ok {
			ret[vg] = d
			continue
		}

		ok, err = prev.GreaterThan(d)
		if err != nil {
			return nil, fmt.Errorf("error parsing version %s: %v", d.ToString(), err)
		} else if ok {
			ret[vg] = d
		}
	}
	return ret, nil
}
