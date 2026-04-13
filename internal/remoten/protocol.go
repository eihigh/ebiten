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
	"encoding/json"
	"errors"
	"fmt"
)

type MessageType string

const (
	MessageHello       MessageType = "hello"
	MessageCommandBatch MessageType = "command_batch"
	MessageInputState  MessageType = "input_state"
	MessageTerminate   MessageType = "terminate"
	MessageError       MessageType = "error"
	MessageBackPressure MessageType = "back_pressure"
)

type Envelope struct {
	Type         MessageType      `json:"type"`
	Hello        *Hello           `json:"hello,omitempty"`
	CommandBatch *CommandBatch    `json:"command_batch,omitempty"`
	InputState   *InputState      `json:"input_state,omitempty"`
	Terminate    *Terminate       `json:"terminate,omitempty"`
	Error        *RemoteError     `json:"error,omitempty"`
	BackPressure *BackPressure    `json:"back_pressure,omitempty"`
}

type Hello struct {
	Role      Role   `json:"role"`
	SessionID string `json:"session_id"`
}

type CommandBatch struct {
	Frame    uint64    `json:"frame"`
	Commands []Command `json:"commands"`
}

type Command struct {
	Op      string          `json:"op"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Shared  *SharedMemoryRef `json:"shared,omitempty"`
}

type SharedMemoryRef struct {
	Name   string `json:"name"`
	Offset uint64 `json:"offset"`
	Size   uint64 `json:"size"`
}

type InputState struct {
	Frame             uint64  `json:"frame"`
	OutsideWidth      float64 `json:"outside_width"`
	OutsideHeight     float64 `json:"outside_height"`
	DeviceScaleFactor float64 `json:"device_scale_factor"`
	Focused           bool    `json:"focused"`
}

type Terminate struct {
	Reason string `json:"reason"`
}

type RemoteError struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

type BackPressure struct {
	QueueDepth   uint32 `json:"queue_depth"`
	MaxQueueDepth uint32 `json:"max_queue_depth"`
}

func (e *Envelope) Validate() error {
	if e == nil {
		return errors.New("remoten: nil envelope")
	}
	if e.Type == "" {
		return errors.New("remoten: empty envelope type")
	}

	var present int
	if e.Hello != nil {
		present++
	}
	if e.CommandBatch != nil {
		present++
	}
	if e.InputState != nil {
		present++
	}
	if e.Terminate != nil {
		present++
	}
	if e.Error != nil {
		present++
	}
	if e.BackPressure != nil {
		present++
	}
	if present != 1 {
		return fmt.Errorf("remoten: envelope type %q requires exactly one payload", e.Type)
	}

	switch e.Type {
	case MessageHello:
		if e.Hello == nil {
			return errors.New("remoten: hello payload is required")
		}
		if e.Hello.Role != RoleHost && e.Hello.Role != RoleGame {
			return fmt.Errorf("remoten: invalid hello role %q", e.Hello.Role)
		}
		if e.Hello.SessionID == "" {
			return errors.New("remoten: hello session id is required")
		}
	case MessageCommandBatch:
		if e.CommandBatch == nil {
			return errors.New("remoten: command_batch payload is required")
		}
		if len(e.CommandBatch.Commands) == 0 {
			return errors.New("remoten: empty command batch")
		}
		for i, c := range e.CommandBatch.Commands {
			if c.Op == "" {
				return fmt.Errorf("remoten: command[%d] has empty op", i)
			}
			if c.Shared != nil {
				if c.Shared.Name == "" {
					return fmt.Errorf("remoten: command[%d] shared memory name is empty", i)
				}
				if c.Shared.Size == 0 {
					return fmt.Errorf("remoten: command[%d] shared memory size is 0", i)
				}
			}
		}
	case MessageInputState:
		if e.InputState == nil {
			return errors.New("remoten: input_state payload is required")
		}
	case MessageTerminate:
		if e.Terminate == nil {
			return errors.New("remoten: terminate payload is required")
		}
	case MessageError:
		if e.Error == nil {
			return errors.New("remoten: error payload is required")
		}
		if e.Error.Message == "" {
			return errors.New("remoten: error message is required")
		}
	case MessageBackPressure:
		if e.BackPressure == nil {
			return errors.New("remoten: back_pressure payload is required")
		}
	default:
		return fmt.Errorf("remoten: unknown envelope type %q", e.Type)
	}
	return nil
}
