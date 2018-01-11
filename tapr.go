// Package tapr includes the main Tapr types and interfaces.
//
// This package MUST NOT import any other Tapr packages.
package tapr // import "hpt.space/tapr"

import (
	"io"
	"time"
)

// A PathName is a string representing a full path name.
// Tapr path names are special. They may match certain pseudo-directories.
//
// Examples
//    /tapr/x/(un)compress/gzip/...
//    /tapr/x/{en,de}crypt/<key>/...
type PathName string

// A Dataset is a collection of files and directories.
type Dataset string

// Config contains client information
type Config interface {
	// Target returns the target of i/o operations.
	Target() string

	// Value returns the value for the given configuration key.
	Value(key string) string
}

// Client is the high-level user API towards Tapr. It is very simplified. The
// client is oblivious to where data is stored, but may give hints.
type Client interface {
	// Pull arranges for the client to pull data from Tapr to an
	// io.Writer.
	Pull(name PathName, w io.Writer) error

	// PullFile is the generalized Pull call. It will pull the named file from
	// the server, starting at offset and writing to w.
	PullFile(name PathName, w io.Writer, offset int64) error

	// Push arranges for the client to push data to Tapr from an
	// io.Reader.
	Push(name PathName, r io.Reader) error

	// PushFile is the generalized Push call. It will push the named file to the
	// server at offset. If append is true, the offset will be ignored.
	PushFile(name PathName, r io.Reader, append bool) error

	// Append appends data from an io.Reader to the named file.
	Append(name PathName, r io.Reader) error

	// Stat retrieves basic file info.
	Stat(name PathName) (*FileInfo, error)
}

// A FileInfo describes a file.
type FileInfo struct {
	Size int64
}

// A NetAddr is the network address of service. It is interpreted by Dialer's
// Dial method to connect to the service.
type NetAddr string

// An Estimate is a time.Duration.
type Estimate time.Duration

// Stager is an interface representing the ability to stage a dataset
type Stager interface {
	Stage(ds Dataset, dst PathName) Estimate
}

// The File interface has semantics and an API that parallels a subset
// of Go's os.File.
type File interface {
	// Close closes an open file.
	Close() error

	// Name returns the full path name of the File.
	Name() string

	// Read, ReadAt, Write, WriteAt and Seek implement
	// the standard Go interfaces io.Reader, etc.
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)

	// Seek sets the offset for the next Read or Write to offset,
	// interpreted according to whence: io.SeekStart means relative
	// to the start of the file, io.SeekCurrent means relative to
	// the current offset, and io.SeekEnd means relative to the end.
	// Seek returns the new offset relative to the start of the file
	// and an error, if any.
	//
	// Seeking to an offset before the start of the file is an error.
	// Seeking to any positive offset is legal, but the behavior of
	// subsequent I/O operations on the underlying object is
	// implementation-dependent.
	Seek(offset int64, whence int) (int64, error)
}
