package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	IstioDistributionFlavorTetrate     = "tetrate"
	IstioDistributionFlavorTetrateFIPS = "tetratefips"
)

func (x *Manifest) GetEOLDates() (map[string]time.Time, error) {
	ret := make(map[string]time.Time, len(x.IstioMinorVersionsEolDates))
	for k, v := range x.IstioMinorVersionsEolDates {
		t, err := parseManifestEOLDate(v)
		if err != nil {
			return nil, err
		}
		ret[k] = t
	}
	return ret, nil
}

func parseManifestEOLDate(in string) (time.Time, error) {
	const layout = "2006-01-02"
	return time.Parse(layout, in)
}

func (x *IstioDistribution) ToString() string {
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

func (x *IstioDistribution) IsOfficial() bool {
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
			x.ToString(), y.ToString(), xg, yg)
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
		// handle the official version schema: 'x.y.z'
		if err := verifyOfficialVersionString(in); err != nil {
			return nil, err
		}

		return &IstioDistribution{Version: in}, nil
	}

	parts := strings.SplitN(in, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid version schema: %s", in)
	}

	if err := verifyOfficialVersionString(parts[0]); err != nil {
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

func verifyOfficialVersionString(in string) error {
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
