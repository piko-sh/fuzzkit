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

package numpool

import (
	"math"
)

const (
	// MantissaBits is the number of bits of a float64's significand below its leading one.
	MantissaBits = 52

	// Float32MantissaBits is that number for a float32.
	Float32MantissaBits = 23

	// ExponentBias is the bias of a float64's exponent.
	ExponentBias = 1023

	// LowestTieBit and HighestTieBit bound the position of the leading bit of an integer
	// TieInteger builds.
	LowestTieBit = 25

	// HighestTieBit is the highest such position.
	HighestTieBit = 62
)

var (
	// FullMantissas lists float64 values whose significands use every bit, so that
	// rounding them to float32 discards set bits.
	FullMantissas = [...]float64{1.0/3.0 + 300e-9, math.Pi, math.E, 0.1, 2.0 / 3.0, math.Sqrt2, 1e10 / 3, -math.Ln2}

	// TieIntegers lists 64-bit integers that a conversion to float32 rounds on or beside
	// a tie.
	TieIntegers = [...]int64{
		1<<60 + 1<<36 + 1, 1<<54 + 1<<30 + 1, -(1<<60 + 1<<36 + 1), -(1<<54 + 1<<30 + 1), 1<<53 + 1, 1<<24 + 1, 1<<62 + 1<<38 + 1<<10,
	}

	// TieUnsigned lists unsigned 64-bit integers of the same kind above the int64 range.
	TieUnsigned = [...]uint64{1<<63 + 1<<39 + 1, 1<<64 - 1, 1<<63 + 1<<39}

	// Boundaries lists float64 values on the boundaries of rounding and remainder
	// functions.
	Boundaries = [...]float64{
		0.49999999999999994, 0.5, -0.5, 1.5, 2.5, -2.5, 3.5, 1<<52 + 1, -(1<<52 + 1), 1<<52 + 3, 1<<53 + 2, 1<<52 + 0.5, 0.1, math.Pi,
		5e-324, 2.2250738585072009e-308, -1e-310, 1e300, math.MaxFloat64, -math.MaxFloat64, 7, -7, 2, 0.75,
	}
)

// TieInteger builds a 64-bit integer whose leading bit is at top and whose conversion to
// float32 rounds on a tie (low is zero) or just past one (low is one).
//
// Takes top (int) which is the position of the leading bit, between LowestTieBit and
// HighestTieBit.
// Takes low (int64) which is the lowest bit, zero or one.
//
// Returns int64 which is the integer.
func TieInteger(top int, low int64) int64 {
	return int64(1)<<top | int64(1)<<(top-Float32MantissaBits-1) | low
}
