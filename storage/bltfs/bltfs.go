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

	"hpt.space/tapr"
	"hpt.space/tapr/errors"
	"hpt.space/tapr/log"
	"hpt.space/tapr/storage"
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
