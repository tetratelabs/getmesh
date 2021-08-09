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
	"os"
	"path/filepath"
)

func extractYamlFilePaths(files []string) (ret []string, err error) {
	for _, f := range files {
		if ext := filepath.Ext(f); ext == ".yaml" {
			ret = append(ret, f)
			continue
		}

		// walk through the directory
		err := filepath.Walk(f,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				if ext := filepath.Ext(path); ext == ".yaml" {
					ret = append(ret, path)
				}
				return nil
			})
		if err != nil {
			return nil, err
		}
	}
	return
}
