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
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func drawSome(source Source) []uint64 {
	var drawn []uint64
	source.Enter(SpanProgram, 1)
	drawn = append(drawn, uint64(source.Weighted([]uint32{1, 0, 5, 3})))
	drawn = append(drawn, source.Uint(10), source.Uint(1000), source.Uint(1<<40), source.Uint(0))
	for source.More(3, 4) && len(drawn) < 64 {
		source.Enter(SpanStatement, uint32(len(drawn)))
		drawn = append(drawn, uint64(source.Pool(7)))
		if source.Bool(1, 2) {
			drawn = append(drawn, 1)
		}
		source.Leave()
	}
	for _, label := range []uint64{9, 9, 4} {
		fork := source.Fork(label)
		drawn = append(drawn, fork.Uint(0), uint64(fork.Pool(300)))
		inner := fork.Fork(1)
		drawn = append(drawn, inner.Uint(256))
	}
	source.Leave()
	return drawn
}

func TestSeededTapeIsDeterministic(t *testing.T) {
	t.Parallel()

	first, second, other := New(42), New(42), New(43)
	require.Equal(t, drawSome(first), drawSome(second))
	require.Equal(t, first.Recording(), second.Recording())
	require.Equal(t, first.Spans(), second.Spans())
	require.NotEqual(t, drawSome(New(42)), drawSome(other), "another seed gives another tape")
}

func TestReplayReproducesEveryDraw(t *testing.T) {
	t.Parallel()

	for _, seed := range []uint64{0, 1, 42, 1 << 63, ^uint64(0)} {
		original := New(seed)
		drawn := drawSome(original)
		recording := original.Recording()

		decoded, err := Unmarshal(recording.Marshal())
		require.NoError(t, err)
		require.Equal(t, recording, decoded)

		replayed := Replay(decoded)
		require.Equal(t, drawn, drawSome(replayed), "seed %d", seed)
		require.Equal(t, recording, replayed.Recording(), "a replay consumes exactly what was recorded")
		require.Equal(t, original.Spans(), replayed.Spans())
		require.Positive(t, recording.Size())
	}
}

func TestForksDoNotDependOnWhatWasDrawnBefore(t *testing.T) {
	t.Parallel()

	early, late := New(7), New(7)
	late.Uint(0)
	late.Uint(100)
	require.Equal(t, early.Fork(3).Uint(0), late.Fork(3).Uint(0), "a declaration's stream depends on the seed and its label only")
	require.NotEqual(t, New(7).Fork(3).Uint(0), New(7).Fork(4).Uint(0))
}

func TestExhaustedTapeSelectsTheSimplestChoice(t *testing.T) {
	t.Parallel()

	for _, recording := range []*Recording{nil, {}, {Data: []byte{}}} {
		empty := Replay(recording)
		require.Equal(t, 2, empty.Weighted([]uint32{0, 0, 4, 9}), "the first index with a positive weight")
		require.Zero(t, empty.Weighted(nil))
		require.Zero(t, empty.Weighted([]uint32{0, 0}))
		require.Zero(t, empty.Uint(10))
		require.Zero(t, empty.Uint(0))
		require.False(t, empty.Bool(255, 256))
		require.False(t, empty.More(9, 10))
		require.Zero(t, empty.Pool(50))
		require.Zero(t, empty.Fork(1).Uint(1000), "a fork of an exhausted tape is exhausted")
	}
}

func TestDrawWidthsAreFixed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		draw  func(tape *Tape)
		width int
	}{
		{name: "a small range is one byte", draw: func(tape *Tape) { tape.Uint(256) }, width: 1},
		{name: "a medium range is two bytes", draw: func(tape *Tape) { tape.Uint(257) }, width: 2},
		{name: "a large range is eight bytes", draw: func(tape *Tape) { tape.Uint(65537) }, width: 8},
		{name: "the whole range is eight bytes", draw: func(tape *Tape) { tape.Uint(0) }, width: 8},
		{name: "a weighted choice is two bytes", draw: func(tape *Tape) { tape.Weighted([]uint32{1, 2}) }, width: 2},
		{name: "a weighted choice without weights draws nothing", draw: func(tape *Tape) { tape.Weighted(nil) }, width: 0},
		{name: "a boolean is one byte", draw: func(tape *Tape) { tape.Bool(1, 2) }, width: 1},
		{name: "a boolean with a zero denominator is still one byte", draw: func(tape *Tape) { tape.Bool(1, 0) }, width: 1},
		{name: "a repetition is one byte", draw: func(tape *Tape) { tape.More(1, 2) }, width: 1},
		{name: "a pool of one still draws", draw: func(tape *Tape) { tape.Pool(1) }, width: 1},
		{name: "a small pool is one byte", draw: func(tape *Tape) { tape.Pool(200) }, width: 1},
		{name: "a large pool is two bytes", draw: func(tape *Tape) { tape.Pool(300) }, width: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tape := New(5)
			tt.draw(tape)
			require.Len(t, tape.Recording().Data, tt.width)
		})
	}
}

func TestDrawsStayInRange(t *testing.T) {
	t.Parallel()

	tape := New(99)
	trueCount := 0
	chosen := make([]int, 3)
	for range 4000 {
		require.Less(t, tape.Uint(7), uint64(7))
		require.Less(t, tape.Uint(1000), uint64(1000))
		require.Less(t, tape.Uint(1<<33), uint64(1<<33))
		pool := tape.Pool(5)
		require.GreaterOrEqual(t, pool, 0)
		require.Less(t, pool, 5)
		if tape.Bool(1, 4) {
			trueCount++
		}
		index := tape.Weighted([]uint32{6, 0, 2})
		chosen[index]++
	}
	require.InDelta(t, 1000, trueCount, 150, "a quarter of the draws are true")
	require.Zero(t, chosen[1], "a zero weight is never chosen")
	require.InDelta(t, 3000, chosen[0], 150)
	require.False(t, New(1).Bool(0, 5), "probability zero is never true")
	require.True(t, New(1).Bool(5, 5), "probability one is always true")
	require.True(t, New(1).Bool(9, 5), "a probability above one is clamped")
}

func TestSpansBracketWhatWasDrawn(t *testing.T) {
	t.Parallel()

	tape := New(3)
	tape.Enter(SpanProgram, 1)
	tape.Uint(0)
	tape.Enter(SpanExpression, 2)
	tape.Uint(10)
	tape.Leave()
	tape.Enter(SpanStatement, 3)
	tape.Leave()
	tape.Leave()
	tape.Leave()
	require.Equal(t, []Span{
		{Start: 0, End: 9, ID: 1, Depth: 0, Kind: SpanProgram},
		{Start: 8, End: 9, ID: 2, Depth: 1, Kind: SpanExpression},
		{Start: 9, End: 9, ID: 3, Depth: 1, Kind: SpanStatement},
	}, tape.Spans(), "an unmatched Leave changes nothing")
}

func TestUnmarshalRejectsMalformedRecordings(t *testing.T) {
	t.Parallel()

	valid := (&Recording{Data: []byte{1, 2, 3}, Forks: []*Recording{{Label: 5, Data: []byte{4}}}}).Marshal()
	deep := &Recording{}
	for range maxRecordingDepth + 2 {
		deep = &Recording{Forks: []*Recording{deep}}
	}
	tests := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: nil},
		{name: "wrong magic", data: []byte("fzt2\x00\x00\x00")},
		{name: "cut short", data: valid[:len(valid)-1]},
		{name: "trailing bytes", data: append(bytes.Clone(valid), 0)},
		{name: "length beyond the data", data: []byte("fzt1\x00\x09\x01")},
		{name: "varint that never ends", data: []byte("fzt1\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff")},
		{name: "too many sub-streams", data: []byte("fzt1\x00\x00\xff\xff\x7f")},
		{name: "nested too deeply", data: deep.Marshal()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := Unmarshal(tt.data)
			require.ErrorIs(t, err, ErrBadRecording)
		})
	}
}

func TestMixSeparatesSeedsAndLabels(t *testing.T) {
	t.Parallel()

	seen := map[uint64]bool{}
	for seed := range uint64(64) {
		for label := range uint64(64) {
			mixed := Mix(seed, label)
			require.False(t, seen[mixed], "Mix(%d, %d) collides", seed, label)
			seen[mixed] = true
		}
	}
	require.Equal(t, Mix(1, 2), Mix(1, 2))
	require.NotEqual(t, Mix(1, 2), Mix(2, 1))
}
