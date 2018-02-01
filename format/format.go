// Package format defines the interfaces that implement storage formats.
package format // import "tapr.space/format"

import (
	"tapr.space/errors"
	"tapr.space/storage"
	"tapr.space/store/tape"
)

// A Constructor is a function that creates a Format.
type Constructor func(cfg tape.FormatConfig) (Formatter, error)

var registration = make(map[string]Constructor)

// Register registers a new Format Constructor with the given name.
func Register(name string, fn Constructor) error {
	const op = "format.Register"
	if _, exists := registration[name]; exists {
		return errors.E(op, errors.Exist)
	}

	registration[name] = fn

	return nil
}

// Formatter is something that can format itself.
type Formatter interface {
	Format(devpath string, serial tape.Serial) (storage.Storage, error)
}

// Mounter specifices that the format can be mounted.
type Mounter interface {
	Mount() error
}

// Unmounter specifices that the format can be unmounted.
type Unmounter interface {
	Unmount() error
}

// Create creates a new storage.Storage implementation using the given format
// type.
func Create(cfg tape.FormatConfig) (Formatter, error) {
	const op = "format.Create"

	fn, found := registration[cfg.Backend]
	if !found {
		return nil, errors.E(op, errors.Invalid, errors.Strf("unknown store backend type: %v", cfg.Backend))
	}

	return fn(cfg)
}
