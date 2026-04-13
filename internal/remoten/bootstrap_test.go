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
	"strings"
	"testing"
)

func TestChildEnv(t *testing.T) {
	base := []string{
		"A=1",
		EnvRole + "=host",
		EnvEndpoint + "=/tmp/old.sock",
		EnvSessionID + "=old",
	}
	env := ChildEnv(base, "/tmp/new.sock", "session")

	foundRole := false
	foundEndpoint := false
	foundSession := false

	for _, kv := range env {
		switch {
		case strings.HasPrefix(kv, EnvRole+"="):
			foundRole = kv == EnvRole+"="+string(RoleGame)
		case strings.HasPrefix(kv, EnvEndpoint+"="):
			foundEndpoint = kv == EnvEndpoint+"=/tmp/new.sock"
		case strings.HasPrefix(kv, EnvSessionID+"="):
			foundSession = kv == EnvSessionID+"=session"
		}
	}
	if !foundRole || !foundEndpoint || !foundSession {
		t.Fatalf("invalid child env: %v", env)
	}
}
