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
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/tetratelabs/getmesh/internal/util/logger"
)

var tableColumns = []string{"NAMESPACE", "NAME", "RESOURCE TYPE", "ERROR CODE", "SEVERITY", "MESSAGE"}

func printResultsWithoutNamespace(results []configValidationResult) {
	formatValidationResults(results)
	data := make([][]string, len(results))
	for i, res := range results {
		data[i] = []string{res.name, res.resourceType, res.errorCode, res.severity.Name, res.message}
	}

	flushTable(tableColumns[1:], data)
}

func printResults(results []configValidationResult) {
	formatValidationResults(results)
	data := make([][]string, len(results))
	for i, res := range results {
		data[i] = []string{res.namespace, res.name, res.resourceType, res.errorCode, res.severity.Name, res.message}
	}

	flushTable(tableColumns, data)
}

func formatValidationResults(in []configValidationResult) {
	for i, r := range in {
		in[i].resourceType = strings.Title(strings.ToLower(r.resourceType))
	}

	sort.Slice(in, func(i, j int) bool {
		return in[i].severity.level < in[j].severity.level
	})
}

func flushTable(columns []string, data [][]string) {
	table := tablewriter.NewWriter(logger.GetWriter())
	table.SetHeader(columns)

	table.SetAutoWrapText(true)
	table.SetColWidth(tablewriter.MAX_ROW_WIDTH * 4)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
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
