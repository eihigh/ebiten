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
	"testing"
)

func TestLoadConfigDisabled(t *testing.T) {
	cfg, err := LoadConfig("linux", func(key string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Enabled {
		t.Fatal("cfg.Enabled must be false")
	}
}

func TestLoadConfigWindowsUnsupported(t *testing.T) {
	cfg, err := LoadConfig("windows", func(key string) string {
		if key == EnvEnabled {
			return "1"
		}
		return ""
	})
	if cfg.Enabled {
		t.Fatal("cfg.Enabled must be false")
	}
	if !errors.Is(err, ErrUnsupportedOnWindows) {
		t.Fatalf("LoadConfig error = %v, want %v", err, ErrUnsupportedOnWindows)
	}
}

func TestLoadConfigGameRoleValidation(t *testing.T) {
	cfg, err := LoadConfig("linux", func(key string) string {
		switch key {
		case EnvEnabled:
			return "true"
		case EnvRole:
			return "game"
		default:
			return ""
		}
	})
	if err == nil {
		t.Fatal("LoadConfig must return an error")
	}
	if cfg.Enabled {
		t.Fatal("cfg.Enabled must be false")
	}
}

func TestLoadConfigLogCategories(t *testing.T) {
	cfg, err := LoadConfig("linux", func(key string) string {
		switch key {
		case EnvEnabled:
			return "1"
		case EnvRole:
			return "host"
		case EnvLogCategories:
			return "ipc, batch, sync, "
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled {
		t.Fatal("cfg.Enabled must be true")
	}
	if cfg.Role != RoleHost {
		t.Fatalf("cfg.Role = %q, want %q", cfg.Role, RoleHost)
	}
	if got, want := len(cfg.LogCategories), 3; got != want {
		t.Fatalf("len(cfg.LogCategories) = %d, want %d", got, want)
	}
}
