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

// +build ignore

// Package bltfs implements a disk backed storage.Storage.
package bltfs

import (
	"os"
	"path/filepath"
	"time"

	"hpt.space/bltfs"
	filedebug "hpt.space/bltfs/backend/file"
	"hpt.space/bltfs/util/fsutil"

	"tapr.space"
	"tapr.space/errors"
	"tapr.space/log"
	"tapr.space/storage"
)

type storageImpl struct {
	store *bltfs.Store
}

var _ storage.Storage = (*storageImpl)(nil)

// New returns a new postgres-backed inventory implementation.
func New(serial, root string) (storage.Storage, error) {
	const op = "storage/disk.New"

	path := filepath.Join(root, serial)

	log.Info.Printf("creating data root: %v", root)

	if err := os.MkdirAll(root, os.ModePerm); err != nil {
		return nil, err
	}

	// setup freshly formated LTFS tape
	if err := fsutil.CopyDir("../bltfs/testdata/ltfs-volume.golden", path); err != nil {
		return nil, err
	}

	backend, err := filedebug.Open(path)
	if err != nil {
		return nil, errors.E(op, err)
	}

	pol := bltfs.RecoveryPolicy{
		FullIndexInterval: 1 * time.Hour,
		DifferentialAfter: 10632560640, // 10 GB
		IncrementalAfter:  1073741824,  // 1 GB
	}

	// open bltfs store
	bltfsStore, err := bltfs.Open(backend,
		// we'll use the file debug backend
		bltfs.WithFileDebug(),

		// set the recovery policy
		bltfs.WithRecoveryPolicy(pol),
	)

	if err != nil {
		return nil, errors.E(op, err)
	}

	return &storageImpl{
		store: bltfsStore,
	}, nil
}

func (stg *storageImpl) Create(path tapr.PathName) (tapr.File, error) {
	log.Info.Printf("storage/disk.Writable: creating file: %v", path)
	return stg.store.Create(string(path))
}

func (stg *storageImpl) Open(path tapr.PathName) (tapr.File, error) {
	return stg.Create(path)
}
