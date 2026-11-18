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

package reduce

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func needs(failing int, required ...int) func(context.Context, []int) bool {
	return func(_ context.Context, sequence []int) bool {
		if len(sequence) == 0 || sequence[len(sequence)-1] != failing {
			return false
		}
		for _, program := range required {
			if !slices.Contains(sequence[:len(sequence)-1], program) {
				return false
			}
		}
		return true
	}
}

func TestSequenceKeepsOnlyTheProgramsTheFailureNeeds(t *testing.T) {
	t.Parallel()

	history := make([]int, 0, 41)
	for program := range 40 {
		history = append(history, program)
	}
	history = append(history, 99)
	require.Equal(t, []int{99}, Sequence(t.Context(), history, needs(99)), "a failure a fresh host shows needs no history")
	require.Equal(t, []int{7, 99}, Sequence(t.Context(), history, needs(99, 7)))
	require.Equal(t, []int{3, 21, 38, 99}, Sequence(t.Context(), history, needs(99, 3, 21, 38)), "the order of the history is kept")
}

func TestSequenceOfAShortHistory(t *testing.T) {
	t.Parallel()

	require.Empty(t, Sequence(t.Context(), []int{}, needs(1)))
	require.Equal(t, []int{1}, Sequence(t.Context(), []int{1}, needs(1)))
	require.Equal(t, []int{4, 5}, Sequence(t.Context(), []int{4, 5}, needs(5, 4)))
}

func TestSequenceStopsWithTheContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	history := []int{1, 2, 3, 4, 9}
	require.Equal(t, history, Sequence(ctx, history, needs(9, 1)), "a cancelled reduction returns what it has")
}
