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

// Package store defines the interfaces that implements the basic
// 'store' abstraction.
package store // import "tapr.space/store"

import (
	"tapr.space/config"
	"tapr.space/errors"
	"tapr.space/storage"
)

// A Constructor is a function that creates a Store.
type Constructor func(name string, cfg config.StoreConfig) (Store, error)

var registration = make(map[string]Constructor)

// Register registers a new Store Constructor with the given name.
func Register(name string, fn Constructor) error {
	const op = "store.Register"
	if _, exists := registration[name]; exists {
		return errors.E(op, errors.Exist)
	}

	registration[name] = fn

	return nil
}

// Store is the store interface.
type Store interface {
	// String returns the target name of the store.
	String() string

	// embed the storage.Storage interface.
	storage.Storage
}

// Create creates a new store using the given named implementation.
func Create(name string, cfg config.StoreConfig) (Store, error) {
	const op = "store.Create"

	fn, found := registration[cfg.Backend]
	if !found {
		return nil, errors.E(op, errors.Invalid, errors.Strf("unknown store backend type: %v", cfg.Backend))
	}

	return fn(name, cfg)
}
