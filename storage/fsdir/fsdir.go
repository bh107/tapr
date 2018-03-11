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

// Package fsdir implements a storage backend that writes to some
// mounted file system.
package fsdir

import (
	"os"
	"path/filepath"

	"tapr.space"
	"tapr.space/storage"
)

type Storage struct {
	root string
}

var _ storage.Storage = (*Storage)(nil)

// New returns a new Storage that reads and writes from the
// specified directory.
func New(path string) *Storage {
	return &Storage{
		root: path,
	}
}

func (s *Storage) Create(name tapr.PathName) (tapr.File, error) {
	return s.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY)
}

func (s *Storage) Open(name tapr.PathName) (tapr.File, error) {
	return s.OpenFile(name, os.O_RDONLY)
}

func (s *Storage) Append(name tapr.PathName) (tapr.File, error) {
	return s.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY)
}

func (s *Storage) OpenFile(name tapr.PathName, flag int) (tapr.File, error) {
	return os.OpenFile(filepath.Join(s.root, string(name)), flag, os.ModePerm)
}

func (s *Storage) Mkdir(name tapr.PathName) error {
	return os.Mkdir(filepath.Join(s.root, string(name)), os.ModePerm)
}

func (s *Storage) MkdirAll(name tapr.PathName) error {
	return os.MkdirAll(filepath.Join(s.root, string(name)), os.ModePerm)
}

func (s *Storage) Stat(name tapr.PathName) (os.FileInfo, error) {
	return os.Stat(filepath.Join(s.root, string(name)))
}
