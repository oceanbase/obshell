/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package stdio

import (
	"bytes"
	"io"
)

// BufferIO simulates an in-memory buffer that implements io.Reader, io.Writer, and io.Closer interfaces.
type BufferIO struct {
	buffer    *bytes.Buffer
	autoClear bool
	closed    bool
}

// NewBufferIO creates a new BufferIO instance.
func NewBufferIO(autoClear bool) *BufferIO {
	return &BufferIO{
		autoClear: autoClear,
		buffer:    bytes.NewBuffer(nil),
	}
}

func (bio *BufferIO) GetBuffer() bytes.Buffer {
	return *bio.buffer
}

// IsTty always returns false to indicate it's not a terminal.
func (bio *BufferIO) IsTTY() bool {
	return false
}

// Writable returns whether the buffer is open for writing.
func (bio *BufferIO) Writable() bool {
	return !bio.closed
}

// Close marks the buffer as closed.
func (bio *BufferIO) Close() error {
	bio.closed = true
	if bio.autoClear {
		bio.Clear()
	}
	return nil
}

// Open resets the buffer and marks it as open.
func (bio *BufferIO) Open() {
	bio.closed = false
	bio.Clear()
}

// Write appends the given bytes to the buffer.
func (bio *BufferIO) Write(p []byte) (n int, err error) {
	if bio.closed {
		return 0, io.ErrClosedPipe
	}
	return bio.buffer.Write(p)
}

// Read reads the next len(p) bytes from the buffer.
func (bio *BufferIO) Read(p []byte) (n int, err error) {
	n, err = bio.buffer.Read(p)
	if err == io.EOF && bio.autoClear {
		bio.Clear()
	}
	return n, err
}

// Clear resets the buffer contents.
func (bio *BufferIO) Clear() {
	bio.buffer.Reset()
}

// Flush clears the buffer if autoClear is enabled.
func (bio *BufferIO) Flush() error {
	if bio.autoClear {
		bio.Clear()
	}
	return nil
}

// String returns the contents of the buffer as a string.
func (bio *BufferIO) String() string {
	return bio.buffer.String()
}
