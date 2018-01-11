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
