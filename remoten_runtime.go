// Copyright 2026 The Ebitengine Authors
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

package ebiten

import (
	"os"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2/internal/remoten"
)

func initializeRemotenRuntime() error {
	cfg, err := remoten.LoadConfig(runtime.GOOS, os.Getenv)
	if err != nil {
		return err
	}
	if !cfg.Enabled {
		return nil
	}

	if cfg.Role == remoten.RoleHost {
		if cfg.SessionID == "" {
			sessionID, err := remoten.NewSessionID()
			if err != nil {
				return err
			}
			if err := os.Setenv(remoten.EnvSessionID, sessionID); err != nil {
				return err
			}
			cfg.SessionID = sessionID
		}
		if cfg.Endpoint == "" {
			if err := os.Setenv(remoten.EnvEndpoint, remoten.DefaultEndpoint(cfg.SessionID)); err != nil {
				return err
			}
		}
	}

	return nil
}
