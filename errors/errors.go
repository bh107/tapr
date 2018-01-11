// Copyright 2016 The Upspin Authors. All rights reserved.
// Copyright 2017 The Tapr Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
//    * Redistributions of source code must retain the above copyright
//      notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
//      copyright notice, this list of conditions and the following
//      disclaimer in the documentation and/or other materials provided
//      with the distribution.
//    * Neither the name of Google Inc. nor the names of its
//      contributors may be used to endorse or promote products derived
//      from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package errors defines the error handling used by corpus.
package errors // import "hpt.space/tapr/errors"

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"runtime"

	"hpt.space/tapr"
	"hpt.space/tapr/log"
)

// Error is the type that implements the error interface.
// It contains a number of fields, each of different type.
// An Error value may leave some values unset.
type Error struct {
	// Path is the Tapr path name of the item being accessed.
	Path tapr.PathName

	// Op is the operation being performed, usually the name of the method
	// being invoked (Get, Put, etc.).
	Op string

	// Kind is the class of error, such as permission failure,
	// or "Other" if its class is unknown or irrelevant.
	Kind Kind

	// The underlying error that triggered this one, if any.
	Err error
}

func (e *Error) isZero() bool {
	return e.Path == "" && e.Op == "" && e.Kind == 0 && e.Err == nil
}

var (
	_ error                      = (*Error)(nil)
	_ encoding.BinaryUnmarshaler = (*Error)(nil)
	_ encoding.BinaryMarshaler   = (*Error)(nil)
)

// Separator is the string used to separate nested errors. By
// default, to make errors easier on the eye, nested errors are
// indented on a new line. A server may instead choose to keep each
// error on a single line by modifying the separator string, perhaps
// to ":: ".
var Separator = ":\n\t"

// Kind defines the kind of error this is, mostly for use by systems
// such as FUSE that must act differently depending on the error.
type Kind uint8

// Kinds of errors.
//
// The values of the error kinds are common between both
// clients and servers. Do not reorder this list or remove
// any items since that will change their values.
// New items must be added only to the end.
const (
	Other      Kind = iota // Unclassified error. This value is not printed in the error message.
	Invalid                // Invalid operation for this type of item.
	Permission             // Permission denied.
	IO                     // External I/O error such as network failure.
	Exist                  // Item already exists.
	NotExist               // Item does not exist.
	IsDir                  // Item is a directory.
	NotDir                 // Item is not a directory..
	NotEmpty               // Directory not empty.
	Private                // Information withheld.
	Internal               // Internal error or inconsistency.
	Transient              // A transient error.
)

func (k Kind) String() string {
	switch k {
	case Other:
		return "other error"
	case Invalid:
		return "invalid operation"
	case Permission:
		return "permission denied"
	case IO:
		return "i/o error"
	case Exist:
		return "item already exists"
	case NotExist:
		return "item does not exist"
	case IsDir:
		return "item is a directory"
	case NotDir:
		return "item is not a directory"
	case NotEmpty:
		return "directory not empty"
	case Private:
		return "information withheld"
	case Internal:
		return "internal error"
	case Transient:
		return "transient error"
	}
	return "unknown error kind"
}

// E builds an error value from its arguments.
// There must be at least one argument or E panics.
// The type of each argument determines its meaning.
// If more than one argument of a given type is presented,
// only the last one is recorded.
//
// The types are:
//	string
//		The operation being performed, usually the method
//		being invoked (Get, Put, etc.)
//	errors.Kind
//		The class of error, such as permission failure.
//	error
//		The underlying error that triggered this one.
//
// If the error is printed, only those items that have been
// set to non-zero values will appear in the result.
//
// If Kind is not specified or Other, we set it to the Kind of
// the underlying error.
//
func E(args ...interface{}) error {
	if len(args) == 0 {
		panic("call to errors.E with no arguments")
	}

	e := &Error{}

	for _, arg := range args {
		switch arg := arg.(type) {
		case tapr.PathName:
			e.Path = arg
		case string:
			e.Op = arg
		case Kind:
			e.Kind = arg
		case *Error:
			// make a copy
			copy := *arg
			e.Err = &copy
		case error:
			e.Err = arg
		default:
			_, file, line, _ := runtime.Caller(1)
			log.Printf("errors.E: bad call from %s:%d: %v", file, line, args)
			return Strf("unknown type %T, value %v in error call", arg, arg)
		}
	}

	// Populate stack information (only in debug mode).
	e.populateStack()

	prev, ok := e.Err.(*Error)
	if !ok {
		return e
	}

	// The previous error was also one of ours. Suppress duplications
	// so the message won't contain the same kind, file name or user name
	// twice.
	if prev.Path == e.Path {
		prev.Path = ""
	}

	if prev.Kind == e.Kind {
		prev.Kind = Other
	}

	// If this error has Kind unset or Other, pull up the inner one.
	if e.Kind == Other {
		e.Kind = prev.Kind
		prev.Kind = Other
	}

	return e
}

// pad appends str to the buffer if the buffer already has some data.
func pad(b *bytes.Buffer, str string) {
	if b.Len() == 0 {
		return
	}

	b.WriteString(str)
}

func (e *Error) Error() string {
	b := new(bytes.Buffer)
	e.printStack(b)

	/*
		if e.Path != "" {
			pad(b, ": ")
			b.WriteString(string(e.Path))
		}

		if e.User != "" {
			if e.Path == "" {
				pad(b, ": ")
			} else {
				pad(b, ", ")
			}

			b.WriteString("user ")
			b.WriteString(string(e.User))
		}
	*/

	if e.Path != "" {
		pad(b, ": ")
		b.WriteString(string(e.Path))
	}

	if e.Op != "" {
		pad(b, ": ")
		b.WriteString(e.Op)
	}
	if e.Kind != 0 {
		pad(b, ": ")
		b.WriteString(e.Kind.String())
	}
	if e.Err != nil {
		// Indent on new line if we are cascading non-empty Upspin errors.
		if prevErr, ok := e.Err.(*Error); ok {
			if !prevErr.isZero() {
				pad(b, Separator)
				b.WriteString(e.Err.Error())
			}
		} else {
			pad(b, ": ")
			b.WriteString(e.Err.Error())
		}
	}
	if b.Len() == 0 {
		return "no error"
	}
	return b.String()
}

// Recreate the errors.New functionality of the standard Go errors package
// so we can create simple text errors when needed.

// Str returns an error that formats as the given text. It is intended to
// be used as the error-typed argument to the E function.
func Str(text string) error {
	return &errorString{text}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

// Strf is equivalent to errors.Errorf, but allows clients to import only this
// package for all error handling.
func Strf(format string, args ...interface{}) error {
	return &errorString{fmt.Sprintf(format, args...)}
}

// MarshalAppend marshals err into a byte slice. The result is appended to b,
// which may be nil.
// It returns the argument slice unchanged if the error is nil.
func (e *Error) MarshalAppend(b []byte) []byte {
	if e == nil {
		return b
	}

	b = appendString(b, string(e.Path))
	b = appendString(b, e.Op)

	var tmp [16]byte
	n := binary.PutVarint(tmp[:], int64(e.Kind))

	b = append(b, tmp[:n]...)
	b = MarshalErrorAppend(e.Err, b)

	return b
}

// MarshalBinary marshals its receiver into a byte slice, which it returns.
// It returns nil if the error is nil. The returned error is always nil.
func (e *Error) MarshalBinary() ([]byte, error) {
	return e.MarshalAppend(nil), nil
}

// MarshalErrorAppend marshals an arbitrary error into a byte slice.
// The result is appended to b, which may be nil.
// It returns the argument slice unchanged if the error is nil.
// If the error is not an *Error, it just records the result of err.Error().
// Otherwise it encodes the full Error struct.
func MarshalErrorAppend(err error, b []byte) []byte {
	if err == nil {
		return b
	}
	if e, ok := err.(*Error); ok {
		// This is an errors.Error. Mark it as such.
		b = append(b, 'E')
		return e.MarshalAppend(b)
	}

	// Ordinary error.
	b = append(b, 'e')
	b = appendString(b, err.Error())

	return b
}

// MarshalError marshals an arbitrary error and returns the byte slice.
// If the error is nil, it returns nil.
// It returns the argument slice unchanged if the error is nil.
// If the error is not an *Error, it just records the result of err.Error().
// Otherwise it encodes the full Error struct.
func MarshalError(err error) []byte {
	return MarshalErrorAppend(err, nil)
}

// UnmarshalBinary unmarshals the byte slice into the receiver, which must be non-nil.
// The returned error is always nil.
func (e *Error) UnmarshalBinary(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	data, b := getBytes(b)
	if data != nil {
		e.Path = tapr.PathName(data)
	}

	data, b = getBytes(b)
	if data != nil {
		e.Op = string(data)
	}

	k, n := binary.Varint(b)
	e.Kind = Kind(k)

	// unmarshal embedded error recursively
	e.Err = UnmarshalError(b[n:])

	return nil
}

// UnmarshalError unmarshals the byte slice into an error value.
// If the slice is nil or empty, it returns nil.
// Otherwise the byte slice must have been created by MarshalError or
// MarshalErrorAppend.
// If the encoded error was of type *Error, the returned error value
// will have that underlying type. Otherwise it will be just a simple
// value that implements the error interface.
func UnmarshalError(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	code := b[0]
	b = b[1:]

	switch code {
	case 'e':
		// plain error
		var data []byte
		data, b = getBytes(b)

		if len(b) != 0 {
			log.Printf("unmarshal error: trailing bytes")
		}

		return Str(string(data))

	case 'E':
		// Error value
		var err Error
		err.UnmarshalBinary(b)

		return &err

	default:
		log.Printf("unmarshal error: corrupt data %q", b)

		return Str(string(b))
	}
}

func appendString(b []byte, str string) []byte {
	var tmp [16]byte
	N := binary.PutUvarint(tmp[:], uint64(len(str)))

	b = append(b, tmp[:N]...)
	b = append(b, str...)

	return b
}

// getBytes unmarshals the byte slice at b (uvarint count followed by bytes)
// and returns the slice followed by the remaining bytes.
// If there is insufficient data, both return values will be nil.
func getBytes(b []byte) (data, remaining []byte) {
	u, n := binary.Uvarint(b)

	if len(b) < n+int(u) {
		log.Printf("unmarshal error: bad encoding")
		return nil, nil
	}

	if n == 0 {
		log.Printf("unmarshal error: bad encoding")
		return nil, b
	}

	return b[n : n+int(u)], b[n+int(u):]
}
