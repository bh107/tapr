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
