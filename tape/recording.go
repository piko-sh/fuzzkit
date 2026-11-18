// Copyright 2026 PolitePixels Limited
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

// This project stands against fascism, authoritarianism, and all forms of
// oppression. We built this to empower people, not to enable those who would
// strip others of their rights and dignity.

package tape

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// recordingMagic starts every serialised recording and carries the format version.
	recordingMagic = "fzt1"

	// maxRecordingDepth bounds the nesting of sub-streams in a serialised recording.
	maxRecordingDepth = 16

	// maxForks bounds the sub-streams of one stream in a serialised recording.
	maxForks = 1 << 16
)

var (
	// ErrBadRecording reports bytes that are not a serialised recording.
	ErrBadRecording = errors.New("tape: malformed recording")
)

// Recording is the consumed tape of one program.
type Recording struct {
	// Data holds the bytes drawn from the stream itself.
	Data []byte

	// Forks holds the sub-streams, in the order they were forked.
	Forks []*Recording

	// Label is the label the stream was forked with.
	//
	// It is zero for the root.
	Label uint64
}

// Marshal serialises the recording.
//
// Returns []byte which Unmarshal reads back.
func (recording *Recording) Marshal() []byte {
	return recording.appendTo([]byte(recordingMagic))
}

// Size returns the number of bytes drawn from the stream and all its sub-streams.
//
// Returns int which is the total.
func (recording *Recording) Size() int {
	total := len(recording.Data)
	for _, fork := range recording.Forks {
		total += fork.Size()
	}
	return total
}

// appendTo appends the stream and its sub-streams to buffer.
//
// Takes buffer ([]byte) which receives the encoding.
//
// Returns []byte which is the extended buffer.
func (recording *Recording) appendTo(buffer []byte) []byte {
	buffer = binary.AppendUvarint(buffer, recording.Label)
	buffer = binary.AppendUvarint(buffer, uint64(len(recording.Data)))
	buffer = append(buffer, recording.Data...)
	buffer = binary.AppendUvarint(buffer, uint64(len(recording.Forks)))
	for _, fork := range recording.Forks {
		buffer = fork.appendTo(buffer)
	}
	return buffer
}

// Unmarshal reads a serialised recording.
//
// Takes data ([]byte) which is what Marshal produced.
//
// Returns *Recording which is the decoded recording.
// Returns error when the bytes are not a well-formed recording.
func Unmarshal(data []byte) (*Recording, error) {
	if len(data) < len(recordingMagic) || string(data[:len(recordingMagic)]) != recordingMagic {
		return nil, fmt.Errorf("%w: missing magic", ErrBadRecording)
	}
	recording, rest, err := decode(data[len(recordingMagic):], 0)
	if err != nil {
		return nil, err
	}
	if len(rest) != 0 {
		return nil, fmt.Errorf("%w: %d trailing bytes", ErrBadRecording, len(rest))
	}
	return recording, nil
}

// decode reads one stream and its sub-streams.
//
// Takes data ([]byte) which starts with the stream's encoding.
// Takes depth (int) which is the nesting depth of the stream.
//
// Returns *Recording which is the decoded stream.
// Returns []byte which is what follows the stream.
// Returns error when the encoding is malformed or nested too deeply.
func decode(data []byte, depth int) (*Recording, []byte, error) {
	if depth > maxRecordingDepth {
		return nil, nil, fmt.Errorf("%w: nested deeper than %d", ErrBadRecording, maxRecordingDepth)
	}
	label, data, err := uvarint(data)
	if err != nil {
		return nil, nil, err
	}
	length, data, err := uvarint(data)
	if err != nil {
		return nil, nil, err
	}
	if length > uint64(len(data)) {
		return nil, nil, fmt.Errorf("%w: stream of %d bytes in %d", ErrBadRecording, length, len(data))
	}
	recording := &Recording{Data: append([]byte(nil), data[:length]...), Forks: nil, Label: label}
	data = data[length:]
	count, data, err := uvarint(data)
	if err != nil {
		return nil, nil, err
	}
	if count > maxForks {
		return nil, nil, fmt.Errorf("%w: %d sub-streams", ErrBadRecording, count)
	}
	for range count {
		var fork *Recording
		fork, data, err = decode(data, depth+1)
		if err != nil {
			return nil, nil, err
		}
		recording.Forks = append(recording.Forks, fork)
	}
	return recording, data, nil
}

// uvarint reads one unsigned varint.
//
// Takes data ([]byte) which starts with the varint.
//
// Returns uint64 which is the value.
// Returns []byte which is what follows it.
// Returns error when the varint is cut short or overflows.
func uvarint(data []byte) (uint64, []byte, error) {
	value, read := binary.Uvarint(data)
	if read <= 0 {
		return 0, nil, fmt.Errorf("%w: bad varint", ErrBadRecording)
	}
	return value, data[read:], nil
}
