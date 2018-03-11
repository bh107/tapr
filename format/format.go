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
	Format(devpath string, vol tape.Volume) (formatted bool, stg storage.Storage, err error)
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
