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
	// splitmixIncrement is the odd constant the generator's state advances by.
	splitmixIncrement = 0x9e3779b97f4a7c15

	// splitmixFirstMultiplier is the first multiplier of the output function.
	splitmixFirstMultiplier = 0xbf58476d1ce4e5b9

	// splitmixSecondMultiplier is the second multiplier of the output function.
	splitmixSecondMultiplier = 0x94d049bb133111eb

	// splitmixFirstShift is the first xor-shift of the output function.
	splitmixFirstShift = 30

	// splitmixSecondShift is the second xor-shift of the output function.
	splitmixSecondShift = 27

	// splitmixFinalShift is the last xor-shift of the output function.
	splitmixFinalShift = 31
)

// splitmix is the splitmix64 generator.
//
// It is the fuzzer's own, so that a program never depends on the generator of a
// standard-library version.
type splitmix struct {
	// state is the generator's whole state.
	state uint64
}

// next returns the next 64 bits.
//
// Returns uint64 which is the next output.
func (generator *splitmix) next() uint64 {
	generator.state += splitmixIncrement
	return scramble(generator.state)
}

// Mix derives one seed from another seed and a label.
//
// Program i of a stream has the seed Mix(base, i), so which programs exist never depends
// on the order they are made in.
//
// Takes seed (uint64) which is the base seed.
// Takes label (uint64) which distinguishes the derived seed.
//
// Returns uint64 which is the derived seed.
func Mix(seed, label uint64) uint64 {
	return scramble(scramble(seed+splitmixIncrement) ^ (label + splitmixIncrement))
}

// scramble is the output function of splitmix64.
//
// Takes value (uint64) which is the state to scramble.
//
// Returns uint64 which is the scrambled value.
func scramble(value uint64) uint64 {
	value = (value ^ (value >> splitmixFirstShift)) * splitmixFirstMultiplier
	value = (value ^ (value >> splitmixSecondShift)) * splitmixSecondMultiplier
	return value ^ (value >> splitmixFinalShift)
}
