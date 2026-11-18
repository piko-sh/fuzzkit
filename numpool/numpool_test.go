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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTieIntegersRoundOnATie(t *testing.T) {
	t.Parallel()

	for top := LowestTieBit; top <= HighestTieBit; top++ {
		number := TieInteger(top, 0)
		above := TieInteger(top, 1)
		if top > MantissaBits {
			require.NotEqual(t, float32(above), float32(float64(above)), "top %d: rounding through float64 lands on the tie", top)
		} else {
			require.Equal(t, float32(above), float32(float64(above)), "top %d: float64 holds the value exactly", top)
		}
		require.Equal(t, float32(number), float32(float64(number)), "top %d: a tie itself is exact in float64", top)
	}
	for _, number := range TieIntegers {
		require.NotZero(t, number)
	}
	for _, number := range TieUnsigned {
		require.Greater(t, number, uint64(math.MaxInt64))
	}
}

func TestPoolsHoldFiniteNumbers(t *testing.T) {
	t.Parallel()

	for _, number := range FullMantissas {
		require.False(t, math.IsInf(number, 0) || math.IsNaN(number))
		require.NotZero(t, math.Float64bits(number)&(1<<(MantissaBits-Float32MantissaBits)-1), "%v loses set bits in float32", number)
	}
	for _, number := range Boundaries {
		require.False(t, math.IsInf(number, 0) || math.IsNaN(number))
	}
	require.Equal(t, uint64(ExponentBias), math.Float64bits(1)>>MantissaBits)
}
