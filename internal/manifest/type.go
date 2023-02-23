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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

type Manifest struct {
	IstioDistributions []*IstioDistribution `json:"istio_distributions"`
	// the end of life of Istio minor versions
	// key: "x.y", "1.7" for example
	// value: "YYYY-MM-DD"
	IstioMinorVersionsEOLDates map[string]string `json:"istio_minor_versions_eol_dates"`
}

type IstioDistribution struct {
	// Distributions are tagged with `x.y.z-${flavor}-v${flavor_version}` where
	// - ${flavor} is either "tetrate" or "tetratefips"
	// - ${flavor_version} is ""numeric"" and the version  of that distribution
	Version       string `json:"version,omitempty"`
	Flavor        string `json:"flavor,omitempty"`
	FlavorVersion int64  `json:"flavor_version,omitempty"`
	// Supported k8s versions of this distribution.
	K8SVersions []string `json:"k8s_versions,omitempty"`
	// Indicates if this is a security update.
	IsSecurityPatch bool `json:"is_security_patch,omitempty"`
	// Release notes for this distribution.
	ReleaseNotes []string `json:"release_notes,omitempty"`
	// EndOfLife of this distribution (format: "YYYY-MM-DD")
	EndOfLife string `json:"end_of_life,omitempty"`
}

const (
	IstioDistributionFlavorTetrate     = "tetrate"
	IstioDistributionFlavorTetrateFIPS = "tetratefips"
	IstioDistributionFlavorIstio       = "istio"
)

func (x *Manifest) GetEOLDates() (map[string]time.Time, error) {
	ret := make(map[string]time.Time, len(x.IstioMinorVersionsEOLDates))
	for k, v := range x.IstioMinorVersionsEOLDates {
		t, err := parseManifestEOLDate(v)
		if err != nil {
			return nil, err
		}
		ret[k] = t
	}
	return ret, nil
}

func (x *Manifest) SetEOLInIstioDistributions() error {
	for _, dist := range x.IstioDistributions {
		iVer, err := semver.NewVersion(dist.Version)
		if err != nil {
			return err
		}
		for v, date := range x.IstioMinorVersionsEOLDates {
			dVer, err := semver.NewVersion(v)
			if err != nil {
				return err
			}
			if (dVer.Major() == iVer.Major()) && (dVer.Minor() == iVer.Minor()) {
				dist.EndOfLife = date
			}
		}
	}
	return nil
}

func parseManifestEOLDate(in string) (time.Time, error) {
	const layout = "2006-01-02"
	return time.Parse(layout, in)
}

func (x *IstioDistribution) String() string {
	return fmt.Sprintf("%s-%s-v%d", x.Version, x.Flavor, x.FlavorVersion)
}

func (x *IstioDistribution) Equal(j *IstioDistribution) bool {
	return x.Version == j.Version &&
		x.Flavor == j.Flavor &&
		x.FlavorVersion == j.FlavorVersion
}

func (x *IstioDistribution) ExistInManifest(ms *Manifest) (bool, error) {
	for _, d := range ms.IstioDistributions {
		if d.Equal(x) {
			return true, nil
		}
	}
	return false, nil
}

func (x *IstioDistribution) Patch() (int, error) {
	ts := strings.Split(x.Version, ".")
	if len(ts) != 3 {
		return 0, fmt.Errorf("invalid vesion: cannot parse %s in the form of 'x.y.z'", x.Version)
	}

	ret, err := strconv.Atoi(ts[2])
	if err != nil {
		return 0, fmt.Errorf("invalid vesion: cannot parse %s: failed to convert %s to int", x.Version, ts[2])
	}
	return ret, nil
}

func (x *IstioDistribution) Group() (string, error) {
	ts := strings.Split(x.Version, ".")
	if len(ts) != 3 {
		return "", fmt.Errorf("invalid vesion: cannot parse %s in the form of 'x.y.z'", x.Version)
	}

	return fmt.Sprintf("%s.%s-%s", ts[0], ts[1], x.Flavor), nil
}

func (x *IstioDistribution) IsUpstream() bool {
	// manifest.json denotes upstream by flavor 'istio'. Whereas the actual upstream images
	// in the cluster is of the form 'x.y.z' with no flavor set
	return x.Flavor == ""
}

// compare two (patch version, flavor version) tuples in the same group
func (x *IstioDistribution) GreaterThan(y *IstioDistribution) (bool, error) {
	xg, err := x.Group()
	if err != nil {
		return false, err
	}
	yg, err := y.Group()
	if err != nil {
		return false, err
	}

	if xg != yg {
		return false, fmt.Errorf("cannot compare two versions (%s, %s) in the differnet group: %s and %s",
			x.String(), y.String(), xg, yg)
	}

	xp, err := x.Patch()
	if err != nil {
		return false, err
	}
	yp, err := y.Patch()
	if err != nil {
		return false, err
	}

	if xp == yp {
		return x.FlavorVersion > y.FlavorVersion, nil
	}

	return xp > yp, nil
}

func IstioDistributionFromString(in string) (*IstioDistribution, error) {
	if !strings.Contains(in, "-") {
		// handle the upstream version schema: 'x.y.z'
		if err := verifyUpstreamVersionString(in); err != nil {
			return nil, err
		}

		return &IstioDistribution{Version: in}, nil
	}

	parts := strings.SplitN(in, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid version schema: %s", in)
	}

	if err := verifyUpstreamVersionString(parts[0]); err != nil {
		return nil, err
	}

	flavor, flavorVersion, err := parseFlavor(parts[1])
	return &IstioDistribution{Version: parts[0], Flavor: flavor, FlavorVersion: flavorVersion}, err
}

func parseFlavor(in string) (string, int64, error) {
	parts := strings.SplitN(in, "-", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid version schema: %s", in)
	}

	flavor := parts[0]
	flavorVersion, err := strconv.ParseInt(strings.TrimPrefix(parts[1], "v"), 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid version schema: %s: %w", in, err)
	}
	return flavor, flavorVersion, nil
}

func verifyUpstreamVersionString(in string) error {
	ts := strings.Split(in, ".")
	if len(ts) != 3 {
		return fmt.Errorf("invalid vesion: cannot parse %s in the form of 'x.y.z'", in)
	}

	for _, t := range ts {
		if _, err := strconv.Atoi(t); err != nil {
			return fmt.Errorf("invalid vesion: cannot parse %s in the form of 'x.y.z'", in)
		}
	}
	return nil
}

// get the istio distribution with latest patch version and latest flavor version
func GetLatestDistribution(current *IstioDistribution, ms *Manifest) (foundLatest *IstioDistribution, includeSecurityPatch bool, err error) {
	tg, err := current.Group()
	if err != nil {
		return nil, false, err
	}

	for _, d := range ms.IstioDistributions {
		dg, err := d.Group()
		if err != nil {
			return nil, false, err
		}

		if tg == dg {
			// if there are any version between current and latest version has security patch
			// includeSecurityPatch should return true
			if ok, _ := d.GreaterThan(current); ok && d.IsSecurityPatch {
				includeSecurityPatch = true
			}

			if foundLatest == nil {
				foundLatest = d
			} else if ok, _ := d.GreaterThan(foundLatest); ok {
				foundLatest = d
			}
		}
	}
	return
}
