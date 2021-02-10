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

package main

import (
	"os"

	// required for authentication against GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/tetratelabs/getistio/cmd"
	"github.com/tetratelabs/getistio/src/getistio"
	"github.com/tetratelabs/getistio/src/util"
	"github.com/tetratelabs/getistio/src/util/logger"
)

var (
	// set by goreleaser
	version = "dev"
)

func main() {
	hd, err := util.GetIstioHomeDir()
	if err != nil {
		logger.Errorf("error initializing getistio home directory: %w", err)
		os.Exit(1)
	}

	if err := getistio.InitConfig(hd); err != nil {
		logger.Errorf("error initializing config: %w", err)
		os.Exit(1)
	}

	cmd.Execute(version, hd)
}
