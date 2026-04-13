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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

const DefaultMaxPayloadSize = 16 << 20

type Framer struct {
	r              io.Reader
	w              io.Writer
	maxPayloadSize int
}

func NewFramer(r io.Reader, w io.Writer, maxPayloadSize int) *Framer {
	if maxPayloadSize <= 0 {
		maxPayloadSize = DefaultMaxPayloadSize
	}
	return &Framer{
		r:              r,
		w:              w,
		maxPayloadSize: maxPayloadSize,
	}
}

func (f *Framer) Send(e *Envelope) error {
	if e == nil {
		return fmt.Errorf("remoten: envelope must not be nil")
	}
	if err := e.Validate(); err != nil {
		return err
	}

	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if len(b) > f.maxPayloadSize {
		return fmt.Errorf("remoten: payload exceeds max size: %d > %d", len(b), f.maxPayloadSize)
	}

	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(b)))
	if _, err := f.w.Write(header[:]); err != nil {
		return err
	}
	if _, err := f.w.Write(b); err != nil {
		return err
	}
	return nil
}

func (f *Framer) Receive() (*Envelope, error) {
	var header [4]byte
	if _, err := io.ReadFull(f.r, header[:]); err != nil {
		return nil, err
	}
	size := int(binary.BigEndian.Uint32(header[:]))
	if size <= 0 || size > f.maxPayloadSize {
		return nil, fmt.Errorf("remoten: invalid payload size: %d", size)
	}

	buf := make([]byte, size)
	if _, err := io.ReadFull(f.r, buf); err != nil {
		return nil, err
	}

	var e Envelope
	if err := json.Unmarshal(buf, &e); err != nil {
		return nil, err
	}
	if err := e.Validate(); err != nil {
		return nil, err
	}
	return &e, nil
}
