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

package util

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

const (
	getmeshDirname     = ".getmesh"
	getmeshHomeEnvName = "GETMESH_HOME" // defaults to ${HOME}/.getmesh
)

func GetmeshHomeDir() (string, error) {
	if home := os.Getenv(getmeshHomeEnvName); len(home) > 0 {
		return home, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return getmeshHomeDir(usr.HomeDir)
}

func getmeshHomeDir(userHome string) (string, error) {
	dir := filepath.Join(userHome, getmeshDirname)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return "", fmt.Errorf("mkdir failed: %v", err)
		}
	} else if err != nil {
		return "", fmt.Errorf("check existence of %s dir: %v", dir, err)
	}
	return dir, nil
}

func HandleMultipleErrors(errorList []error) error {
	if len(errorList) == 0 {
		return nil
	}

	var toPrintErrorCollection string

	for _, err := range errorList {
		toPrintErrorCollection = fmt.Sprintf("%s\n%s", toPrintErrorCollection, err)
	}

	toPrintErrorCollection = fmt.Sprintf("%s\n", toPrintErrorCollection)

	return errors.New(toPrintErrorCollection)
}
