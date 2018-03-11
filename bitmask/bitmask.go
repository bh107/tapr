// Copyright 2018 Klaus Birkelund Abildgaard Jensen
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

// Package bitmask implements simple functions for handling
// bitmasks/flags.
package bitmask

// IsSet returns true if the given flag is set.
func IsSet(m, flag uint32) bool { return m&flag != 0 }

// Set sets the flag in the mask.
func Set(m *uint32, flag uint32) { *m |= flag }

// Clear clears the flag in the mask.
func Clear(m *uint32, flag uint32) { *m &= ^flag }

// Toggle toggles the flag in the mask.
func Toggle(m *uint32, flag uint32) { *m ^= flag }
