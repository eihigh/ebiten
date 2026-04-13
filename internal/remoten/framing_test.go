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
	"bytes"
	"strings"
	"testing"
)

func TestFramerRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	fw := NewFramer(nil, &buf, 0)
	fr := NewFramer(&buf, nil, 0)

	env := &Envelope{
		Type: MessageCommandBatch,
		CommandBatch: &CommandBatch{
			Frame: 1,
			Commands: []Command{
				{Op: "draw_triangles"},
			},
		},
	}
	if err := fw.Send(env); err != nil {
		t.Fatal(err)
	}

	got, err := fr.Receive()
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != MessageCommandBatch {
		t.Fatalf("got.Type = %q, want %q", got.Type, MessageCommandBatch)
	}
	if got.CommandBatch == nil || len(got.CommandBatch.Commands) != 1 {
		t.Fatal("invalid command batch")
	}
}

func TestFramerTooLarge(t *testing.T) {
	var buf bytes.Buffer
	f := NewFramer(nil, &buf, 32)
	env := &Envelope{
		Type: MessageError,
		Error: &RemoteError{
			Kind:    "test",
			Message: strings.Repeat("x", 128),
		},
	}
	if err := f.Send(env); err == nil {
		t.Fatal("Send must return an error")
	}
}

func TestEnvelopeValidate(t *testing.T) {
	env := &Envelope{
		Type:         MessageInputState,
		CommandBatch: &CommandBatch{Frame: 1, Commands: []Command{{Op: "x"}}},
	}
	if err := env.Validate(); err == nil {
		t.Fatal("Validate must return an error for mismatched payload")
	}
}
