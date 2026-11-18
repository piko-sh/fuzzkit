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
)

// Sequence shrinks the items preceding a failure that depends on earlier items, while
// keeping the failing item last.
//
// Takes history ([]T) which holds the items in the order they ran, the failing one
// last.
// Takes fails (func(context.Context, []T) bool) which runs a sequence from a fresh start
// and reports whether its last item still fails as it did.
//
// Returns []T which is the shortest sequence found, the failing item last. The whole
// history when the context ends first or the history is empty.
func Sequence[T any](ctx context.Context, history []T, fails func(context.Context, []T) bool) []T {
	if len(history) <= 1 {
		return history
	}
	failing := history[len(history)-1]
	with := func(prefix []T) []T {
		return append(append(make([]T, 0, len(prefix)+1), prefix...), failing)
	}
	prefix := append([]T(nil), history[:len(history)-1]...)
	if fails(ctx, with(nil)) {
		return with(nil)
	}
	granularity := 2
	for len(prefix) > 0 && ctx.Err() == nil {
		size := (len(prefix) + granularity - 1) / granularity
		reduced := false
		for start := 0; start < len(prefix) && ctx.Err() == nil; start += size {
			end := min(start+size, len(prefix))
			complement := append(append(make([]T, 0, len(prefix)-(end-start)), prefix[:start]...), prefix[end:]...)
			if fails(ctx, with(complement)) {
				prefix = complement
				granularity = max(granularity-1, 2)
				reduced = true
				break
			}
		}
		if reduced {
			continue
		}
		if granularity >= len(prefix) {
			break
		}
		granularity = min(granularity*2, len(prefix))
	}
	return with(prefix)
}
