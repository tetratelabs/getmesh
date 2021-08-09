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

package manifestchecker

import (
	"strings"
	"time"

	"github.com/Masterminds/semver"

	"github.com/tetratelabs/getmesh/src/getmesh"
	"github.com/tetratelabs/getmesh/src/manifest"
	"github.com/tetratelabs/getmesh/src/util/logger"
)

func endOfLifeChecker(m *manifest.Manifest) error {
	return endOfLifeCheckerImpl(m, time.Now())
}

func endOfLifeCheckerImpl(m *manifest.Manifest, now time.Time) error {
	current := getmesh.GetActiveConfig().IstioDistribution
	if current == nil {
		return nil
	}

	currentVer, err := semver.NewVersion(current.Version)
	if err != nil {
		return err
	}

	dates, err := m.GetEOLDates()
	if err != nil {
		return err
	}

	var greaterVersions []string
	for _, d := range m.IstioDistributions {
		v, err := semver.NewVersion(d.Version)
		if err != nil {
			return err
		}

		if v.Minor() > currentVer.Minor() {
			greaterVersions = append(greaterVersions, d.String())
		}

	}

	for mv, eol := range dates {
		v, err := semver.NewVersion(mv)
		if err != nil {
			return err
		}

		if v.Minor() == currentVer.Minor() && eol.UTC().AddDate(0, -1, 0).Before(now) {
			logger.Warnf("Your current active minor version %s is reaching the end of life on %s. "+
				"We strongly recommend you to upgrade to the available higher minor versions: %s.\n",
				mv, eol.Format("2006-01-02"), strings.Join(greaterVersions, ", "))
		}
	}

	return nil
}
