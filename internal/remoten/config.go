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

package remoten

import (
	"errors"
	"fmt"
	"strings"
)

const (
	EnvEnabled       = "EBITENGINE_REMOTEN"
	EnvRole          = "EBITENGINE_REMOTEN_ROLE"
	EnvEndpoint      = "EBITENGINE_REMOTEN_ENDPOINT"
	EnvSessionID     = "EBITENGINE_REMOTEN_SESSION"
	EnvLogCategories = "EBITENGINE_REMOTEN_LOG"
)

type Role string

const (
	RoleHost Role = "host"
	RoleGame Role = "game"
)

type Config struct {
	Enabled       bool
	Role          Role
	Endpoint      string
	SessionID     string
	LogCategories []string
}

var ErrUnsupportedOnWindows = errors.New("ebiten: remoten is not supported on Windows yet")

func LoadConfig(goos string, getenv func(string) string) (Config, error) {
	cfg := Config{}

	if !parseBool(getenv(EnvEnabled)) {
		return cfg, nil
	}
	if goos == "windows" {
		cfg.Enabled = false
		return cfg, ErrUnsupportedOnWindows
	}

	cfg.Enabled = true
	cfg.Role = RoleHost

	if role := strings.TrimSpace(getenv(EnvRole)); role != "" {
		cfg.Role = Role(strings.ToLower(role))
	}
	if cfg.Role != RoleHost && cfg.Role != RoleGame {
		return Config{}, fmt.Errorf("ebiten: invalid %s: %q", EnvRole, cfg.Role)
	}

	cfg.Endpoint = strings.TrimSpace(getenv(EnvEndpoint))
	cfg.SessionID = strings.TrimSpace(getenv(EnvSessionID))

	if cfg.Role == RoleGame {
		if cfg.Endpoint == "" {
			return Config{}, fmt.Errorf("ebiten: %s must be set when %s=%s", EnvEndpoint, EnvRole, RoleGame)
		}
		if cfg.SessionID == "" {
			return Config{}, fmt.Errorf("ebiten: %s must be set when %s=%s", EnvSessionID, EnvRole, RoleGame)
		}
	}

	if raw := strings.TrimSpace(getenv(EnvLogCategories)); raw != "" {
		for _, c := range strings.Split(raw, ",") {
			c = strings.TrimSpace(c)
			if c == "" {
				continue
			}
			cfg.LogCategories = append(cfg.LogCategories, c)
		}
	}

	return cfg, nil
}

func parseBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "t", "yes", "y", "on":
		return true
	default:
		return false
	}
}
