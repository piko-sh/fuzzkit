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

// Package tape is a replayable source of choices for test-case generators.
//
// A tape is a stream of bytes read through draws of fixed width. A seeded generator is
// one way of filling it and a recorded byte string is another. Nothing that draws can
// tell which is in use. An exhausted tape yields zeros, and a zero draw always selects
// the simplest choice, so every byte string, the empty one included, denotes a test
// case.
package tape
