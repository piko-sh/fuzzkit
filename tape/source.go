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

const (
	// SpanProgram brackets a whole program.
	SpanProgram SpanKind = iota

	// SpanDeclaration brackets one top-level declaration.
	SpanDeclaration

	// SpanStatement brackets one statement.
	SpanStatement

	// SpanExpression brackets one expression production.
	SpanExpression
)

// SpanKind says what a span of the tape produced.
type SpanKind uint8

// Source is what the generator draws from.
type Source interface {
	// Weighted draws an index in proportion to positive weights.
	//
	// A zero draw chooses the first positive weight. No positive weights returns zero.
	//
	// Takes weights ([]uint32) which assigns a weight to each choice.
	//
	// Returns int which is the selected weight index, or zero when all weights are zero.
	Weighted(weights []uint32) int

	// Uint returns a value below limit.
	//
	// A zero draw selects zero.
	//
	// Takes limit (uint64) which is the exclusive upper bound.
	//
	// Returns uint64 which is a value below limit.
	Uint(limit uint64) uint64

	// Bool is true with probability numerator over denominator.
	//
	// A zero draw selects false.
	//
	// Takes numerator (uint32) which sets the accepted share.
	// Takes denominator (uint32) which sets the total share.
	//
	// Returns bool which is the drawn value with probability numerator over denominator.
	Bool(numerator, denominator uint32) bool

	// More answers "one more element?" and is drawn before each element.
	//
	// A zero draw selects false.
	//
	// Takes numerator (uint32) which sets the accepted share.
	// Takes denominator (uint32) which sets the total share.
	//
	// Returns bool which reports whether to include another element.
	More(numerator, denominator uint32) bool

	// Pool returns an index into a pool of the given size.
	//
	// A zero draw selects the first, plainest entry.
	//
	// Takes size (int) which is the number of pool entries.
	//
	// Returns int which is an index into a pool of the given size.
	Pool(size int) int

	// Enter opens a span, recorded beside the tape.
	//
	// Takes kind (SpanKind) which identifies the span category.
	// Takes id (uint32) which identifies the span.
	Enter(kind SpanKind, id uint32)

	// Leave closes the innermost open span.
	Leave()

	// Fork returns the sub-stream of one top-level declaration.
	//
	// Takes label (uint64) which names the tape region.
	//
	// Returns Source which is the sub-stream of one top-level declaration.
	Fork(label uint64) Source
}

// Span is one recorded bracket of the tape.
type Span struct {
	// Start is the offset of the first byte the span consumed.
	Start int

	// End is the offset after the last byte the span consumed.
	End int

	// ID identifies what was produced.
	//
	// Its meaning depends on the kind.
	ID uint32

	// Depth is the nesting depth of the span.
	Depth uint16

	// Kind says what the span produced.
	Kind SpanKind
}
