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
	"math"
)

const (
	// narrowWidth is the width, in bytes, of a draw over a small range.
	narrowWidth = 1

	// weightedWidth is the width of a weighted draw.
	weightedWidth = 2

	// wideWidth is the width of a raw scalar draw.
	wideWidth = 8

	// narrowRange is the number of values a narrow draw distinguishes.
	narrowRange = 1 << 8

	// weightedRange is the number of values a weighted draw distinguishes.
	weightedRange = 1 << 16

	// byteBits is the number of bits in a byte.
	byteBits = 8
)

// Tape is a Source that records everything drawn from it.
//
// It is filled either by a seeded generator or by a recording made earlier.
type Tape struct {
	// replay is the recording being replayed.
	//
	// It is nil when the tape is filled from a seed.
	replay *Recording

	// data holds every byte drawn so far.
	data []byte

	// forks holds the sub-streams, in the order they were forked.
	forks []*Tape

	// spans holds the closed and open spans, in the order they were opened.
	spans []Span

	// open holds the indices into spans of the spans not yet closed.
	open []int

	// generator fills the tape when it is not a replay.
	generator splitmix

	// seed is what the generator was seeded with.
	//
	// Sub-streams derive theirs from it.
	seed uint64

	// label is the label the tape was forked with.
	label uint64
}

// New returns a tape filled by the fuzzer's generator.
//
// Takes seed (uint64) which seeds the generator.
//
// Returns *Tape which records what is drawn.
func New(seed uint64) *Tape {
	return &Tape{replay: nil, data: nil, forks: nil, spans: nil, open: nil, generator: splitmix{state: seed}, seed: seed, label: 0}
}

// Replay returns a tape that yields the bytes of a recording, then zeros.
//
// Takes recording (*Recording) which is the recording to replay. A nil value replays the
// empty tape.
//
// Returns *Tape which records what is drawn, so a replay can be compared with its origin.
func Replay(recording *Recording) *Tape {
	if recording == nil {
		recording = &Recording{Data: nil, Forks: nil, Label: 0}
	}
	return &Tape{replay: recording, data: nil, forks: nil, spans: nil, open: nil, generator: splitmix{state: 0}, seed: 0, label: recording.Label}
}

// Weighted implements Source.
//
// Takes weights ([]uint32) which holds one weight per choice.
//
// Returns int which is the chosen index.
func (tape *Tape) Weighted(weights []uint32) int {
	var total uint64
	for _, weight := range weights {
		total += uint64(weight)
	}
	if total == 0 {
		return 0
	}
	target := tape.draw(weightedWidth) * total / weightedRange
	for index, weight := range weights {
		if target < uint64(weight) {
			return index
		}
		target -= uint64(weight)
	}
	return len(weights) - 1
}

// Uint implements Source.
//
// Takes limit (uint64) which is the exclusive upper bound. Zero means the whole range.
//
// Returns uint64 which is below limit.
func (tape *Tape) Uint(limit uint64) uint64 {
	switch {
	case limit == 0:
		return tape.draw(wideWidth)
	case limit <= narrowRange:
		return tape.draw(narrowWidth) % limit
	case limit <= weightedRange:
		return tape.draw(weightedWidth) % limit
	default:
		return tape.draw(wideWidth) % limit
	}
}

// Bool implements Source.
//
// Takes numerator (uint32) which is the numerator of the probability of true.
// Takes denominator (uint32) which is its denominator.
//
// Returns bool which is true for the highest draws, so that a zero draw is false.
func (tape *Tape) Bool(numerator, denominator uint32) bool {
	drawn := tape.draw(narrowWidth)
	if denominator == 0 {
		return false
	}
	threshold := uint64(narrowRange) * uint64(min(numerator, denominator)) / uint64(denominator)
	return drawn >= narrowRange-threshold
}

// More implements Source.
//
// Takes numerator (uint32) which is the numerator of the probability of another element.
// Takes denominator (uint32) which is its denominator.
//
// Returns bool which is true when another element follows.
func (tape *Tape) More(numerator, denominator uint32) bool {
	return tape.Bool(numerator, denominator)
}

// Pool implements Source.
//
// Takes size (int) which is the size of the pool.
//
// Returns int which is an index into the pool, zero for a pool of one or none.
func (tape *Tape) Pool(size int) int {
	if size <= 1 {
		_ = tape.draw(narrowWidth)
		return 0
	}
	drawn := tape.Uint(uint64(size))
	if drawn > math.MaxInt32 {
		return 0
	}
	return int(drawn)
}

// Enter implements Source.
//
// Takes kind (SpanKind) which says what the span produces.
// Takes id (uint32) which identifies it.
func (tape *Tape) Enter(kind SpanKind, id uint32) {
	depth := uint16(min(len(tape.open), math.MaxUint16))
	tape.open = append(tape.open, len(tape.spans))
	tape.spans = append(tape.spans, Span{Start: len(tape.data), End: len(tape.data), ID: id, Depth: depth, Kind: kind})
}

// Leave implements Source.
//
// A Leave without a matching Enter does nothing.
func (tape *Tape) Leave() {
	if len(tape.open) == 0 {
		return
	}
	last := len(tape.open) - 1
	tape.spans[tape.open[last]].End = len(tape.data)
	tape.open = tape.open[:last]
}

// Fork implements Source.
//
// A seeded tape derives the sub-stream's seed from its own seed and the label, so the
// sub-stream does not depend on how much was drawn before the fork. A replay hands out
// the recorded sub-streams in the order they were forked.
//
// Takes label (uint64) which names the declaration the sub-stream belongs to.
//
// Returns Source which is the sub-stream.
func (tape *Tape) Fork(label uint64) Source {
	var child *Tape
	if tape.replay == nil {
		child = New(Mix(tape.seed, label))
	} else {
		var recorded *Recording
		if index := len(tape.forks); index < len(tape.replay.Forks) {
			recorded = tape.replay.Forks[index]
		}
		child = Replay(recorded)
	}
	child.label = label
	tape.forks = append(tape.forks, child)
	return child
}

// Spans returns the spans recorded so far, in the order they were opened.
//
// Returns []Span which is the tape's own slice. The caller must not change it.
func (tape *Tape) Spans() []Span {
	return tape.spans
}

// Recording returns everything drawn from the tape and its sub-streams.
//
// Returns *Recording which shares nothing with the tape.
func (tape *Tape) Recording() *Recording {
	recording := &Recording{Data: append([]byte(nil), tape.data...), Forks: nil, Label: tape.label}
	if len(tape.forks) > 0 {
		recording.Forks = make([]*Recording, 0, len(tape.forks))
	}
	for _, fork := range tape.forks {
		recording.Forks = append(recording.Forks, fork.Recording())
	}
	return recording
}

// draw reads width bytes, little-endian, and records them.
//
// Takes width (int) which is the number of bytes to read.
//
// Returns uint64 which is the value read. Zero once a replayed tape is exhausted.
func (tape *Tape) draw(width int) uint64 {
	var value uint64
	if tape.replay == nil {
		value = tape.generator.next()
		if width < wideWidth {
			value &= 1<<(width*byteBits) - 1
		}
	} else {
		offset := len(tape.data)
		for index := range width {
			if offset+index < len(tape.replay.Data) {
				value |= uint64(tape.replay.Data[offset+index]) << (index * byteBits)
			}
		}
	}
	var encoded [wideWidth]byte
	binary.LittleEndian.PutUint64(encoded[:], value)
	tape.data = append(tape.data, encoded[:width]...)
	return value
}
