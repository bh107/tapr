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

package storage // import "tapr.space/storage"

import (
	"os"

	"tapr.space"
)

// Storage is the storage interface.
type Storage interface {
	// Create creates the named file, truncating it if it already exists.
	Create(tapr.PathName) (tapr.File, error)

	// Open opens the named file for reading.
	Open(tapr.PathName) (tapr.File, error)

	// OpenFile is the generalized open call. It opens the named file with
	// the specified flag.
	OpenFile(name tapr.PathName, flag int) (tapr.File, error)

	// Append opens the named file for appending.
	Append(tapr.PathName) (tapr.File, error)

	// Stat returns a FileInfo describing the named file.
	Stat(tapr.PathName) (os.FileInfo, error)

	// Mkdir creates a new directory with the specified name and permission bits.
	Mkdir(tapr.PathName) error

	// MkdirAll creates a directory named path, along with any necessary parents,
	// and returns nil, or else returns an error. The permission bits perm are
	// used for all directories that MkdirAll creates. If path is already a
	// directory, MkdirAll does nothing and returns nil.
	MkdirAll(tapr.PathName) error
}
