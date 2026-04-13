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
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func NewSessionID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func DefaultEndpoint(sessionID string) string {
	return filepath.Join(os.TempDir(), "ebitengine-remoten-"+sessionID+".sock")
}

func ChildEnv(base []string, endpoint, sessionID string) []string {
	filtered := make([]string, 0, len(base)+4)
	for _, kv := range base {
		if strings.HasPrefix(kv, EnvRole+"=") ||
			strings.HasPrefix(kv, EnvEndpoint+"=") ||
			strings.HasPrefix(kv, EnvSessionID+"=") {
			continue
		}
		filtered = append(filtered, kv)
	}
	filtered = append(filtered,
		EnvEnabled+"=1",
		EnvRole+"="+string(RoleGame),
		EnvEndpoint+"="+endpoint,
		EnvSessionID+"="+sessionID,
	)
	return filtered
}

func StartGameProcess(ctx context.Context, executable string, args []string, env []string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}
