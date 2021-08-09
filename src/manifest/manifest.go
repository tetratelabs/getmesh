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

package manifest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/olekukonko/tablewriter"

	"github.com/tetratelabs/getmesh/src/util/logger"
)

const (
	manifestURL = "https://istio.tetratelabs.io/getmesh/manifest.json"
)

// GlobalManifestURLMux for test purpose
var GlobalManifestURLMux sync.Mutex

func FetchManifest() (ret *Manifest, err error) {
	if p := os.Getenv("GETMESH_TEST_MANIFEST_PATH"); len(p) != 0 {
		raw, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(raw, &ret); err != nil {
			return nil, fmt.Errorf("error unmarshalling fetched manifest: %v", err)
		}
	} else {
		ret, err = fetchManifest(manifestURL)
		if err != nil {
			return nil, err
		}
	}
	return
}

func fetchManifest(url string) (*Manifest, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching manifest: %v", err)
	}

	defer res.Body.Close()
	raw, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading fetched manifest: %v ", err)
	}

	var ret Manifest
	if err := json.Unmarshal(raw, &ret); err != nil {
		return nil, fmt.Errorf("error unmarshalling fetched manifest: %v", err)
	}

	return &ret, nil
}

func PrintManifest(ms *Manifest, current *IstioDistribution) error {
	column := []string{"ISTIO VERSION", "FLAVOR", "FLAVOR VERSION", "K8S VERSIONS"}
	data := make([][]string, len(ms.IstioDistributions))
	for i, m := range ms.IstioDistributions {
		ps := strings.Join(m.K8SVersions, ",")
		if current != nil && m.Equal(current) {
			m.Version = "*" + m.Version
		}
		data[i] = []string{m.Version, m.Flavor,
			strconv.Itoa(int(m.FlavorVersion)), ps}
	}

	table := tablewriter.NewWriter(logger.GetWriter())
	table.SetHeader(column)
	flushTable(table, data)
	return nil
}

func flushTable(table *tablewriter.Table, data [][]string) {
	table.SetAutoWrapText(true)
	table.SetColWidth(tablewriter.MAX_ROW_WIDTH * 4)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data) // Add Bulk Data
	table.Render()
}
