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

package getistio

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/tetratelabs/getistio/api"
)

// GlobalConfigMux for test purpose
var GlobalConfigMux sync.Mutex

type Config struct {
	IstioDistribution *api.IstioDistribution `json:"istio_distribution"`
	DefaultHub        string                 `json:"default_hub,omitempty"`
}

var currentConfig Config

// for switch
func SetIstioVersion(homedir string, d *api.IstioDistribution) error {
	configPath := getConfigPath(homedir)
	currentConfig.IstioDistribution = d
	raw, err := json.Marshal(currentConfig)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}
	if err := ioutil.WriteFile(configPath, raw, 0644); err != nil {
		return fmt.Errorf("error writing configuration at %s: %v", configPath, err)
	}
	return nil
}

// for default-hub
func SetDefaultHub(homedir, hub string) error {
	configPath := getConfigPath(homedir)
	currentConfig.DefaultHub = hub
	raw, err := json.Marshal(currentConfig)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}
	if err := ioutil.WriteFile(configPath, raw, 0644); err != nil {
		return fmt.Errorf("error writing configuration at %s: %v", configPath, err)
	}
	return nil
}

// for istio cmd
func GetActiveConfig() Config {
	return currentConfig
}

func InitConfig(homedir string) error {
	configPath := getConfigPath(homedir)
	_, err := os.Stat(configPath)
	if err == nil {
		raw, err := ioutil.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("read configuration file at %s: %v", configPath, err)
		}
		if err := json.Unmarshal(raw, &currentConfig); err != nil {
			return fmt.Errorf("error unmarshalling configuration for %s: %v", configPath, err)
		}
		return nil
	} else if !os.IsNotExist(err) && err != nil {
		return fmt.Errorf("failed to open configuration file at %s: %v", configPath, err)
	}

	currentConfig = Config{}
	raw, err := json.Marshal(currentConfig)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	if err := ioutil.WriteFile(configPath, raw, 0644); err != nil {
		return fmt.Errorf("error writing configuration at %s: %v", configPath, err)
	}
	return nil
}

func getConfigPath(homedir string) string {
	const name = "config.json"
	return filepath.Join(homedir, name)
}
