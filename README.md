# fuzzkit

Building blocks for generators, fuzzers and property tests in Go: a replayable source of
choices, pools of values at numeric rounding boundaries, and a sequence minimiser.

Module `piko.sh/fuzzkit`. Go 1.27. Apache-2.0.

## Install

```sh
go get piko.sh/fuzzkit
```

## Usage

```go
package convert_test

import (
	"testing"

	"piko.sh/fuzzkit/numpool"
)

// toFloat32 is the code under test: it must round once, as Go's conversion does.
func toFloat32(x int64) float32 { return float32(float64(x)) }

func TestToFloat32RoundsOnce(t *testing.T) {
	for top := numpool.LowestTieBit; top <= numpool.HighestTieBit; top++ {
		for low := int64(0); low <= 1; low++ {
			x := numpool.TieInteger(top, low)
			if got, want := toFloat32(x), float32(x); got != want {
				t.Errorf("toFloat32(%d) = %v, want %v", x, got, want)
			}
		}
	}
}
```

Rounding through `float64` first gives a different result for integers just past a
float32 tie, which the pool finds and random inputs rarely do.

## What it provides

- **`tape`** is a replayable source of choices. **`New(seed)`** fills a tape from a
  seeded generator and **`Replay(recording)`** from a recorded byte string; code that
  draws (**`Uint`**, **`Weighted`**, **`Bool`**, **`More`**, **`Pool`**) cannot tell
  which. Every draw has a fixed width, an exhausted tape yields zeros, and a zero draw
  always selects the simplest choice, so every byte string, the empty one included,
  denotes a test case. **`Recording`** captures what was drawn; **`Marshal`** and
  **`Unmarshal`** store it. **`Enter`** and **`Leave`** record spans beside the tape,
  and **`Fork`** gives an independent sub-stream whose draws do not depend on what was
  drawn before it.
- **`numpool`** lists values near numeric rounding and conversion boundaries:
  **`TieIntegers`** and **`TieUnsigned`** (64-bit integers whose conversion to float32
  rounds on or beside a tie), **`TieInteger(top, low)`** to build one with its leading bit
  at any position, **`FullMantissas`** (float64 values whose rounding to float32 discards
  set bits) and **`Boundaries`** (float64 values on the boundaries of rounding and
  remainder functions).
- **`reduce`** has **`Sequence`**, which shrinks the items preceding a failure that depends
  on earlier items while keeping the failing item last, using delta debugging.

No function starts a goroutine. A `Tape` is not safe for concurrent use.

## Development

```sh
make test       # run the test suite
make test-race  # run it with the race detector
```

Run `make help` for all development commands.
