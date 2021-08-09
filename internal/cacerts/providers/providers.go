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

package providers

import (
	"context"

	"github.com/tetratelabs/getmesh/internal/cacerts/k8s"
	"github.com/tetratelabs/getmesh/internal/cacerts/providers/models"
)

// ProviderInterface defines the operations available on cloud providers
type ProviderInterface interface {
	// IssueCA creates an new intermediate CA according to the provider specified
	// and returns the CA File Path and error (if any)
	IssueCA(context.Context, models.IssueCAOptions) (*k8s.IstioSecretDetails, error)
}
