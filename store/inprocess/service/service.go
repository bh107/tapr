// +build ignore

// Package service implements a simple in-memory non-persistent store.Store.
package service

import (
	"bytes"
	"os"
	"sync"
	"time"

	"hpt.space/tapr"
	"hpt.space/tapr/store"
)

type service struct {
	name string
	data *dataService
}

type dataService struct {
	mu struct {
		sync.Mutex
		blobs map[tapr.PathName][]byte
	}
}

var _ store.Store = (*service)(nil)

// New creates a new inprocess storage.Storage service.
func New() store.Store {
	svc := &service{
		data: &dataService{},
	}

	svc.data.mu.blobs = make(map[tapr.PathName][]byte)

	return svc
}

type file struct {
	name tapr.PathName
	buf  *bytes.Buffer
}

func (s *service) String() string {
	return s.name
}

func (f *file) Close() error {
	return nil
}

func (f *file) Name() string {
	return string(f.name)
}

func (f *file) Read(p []byte) (n int, err error) {
	return f.buf.Read(p)
}

func (f *file) Write(p []byte) (n int, err error) {
	return f.buf.Write(p)
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	panic("not implemented")
}

func (s *service) Create(name tapr.PathName) (tapr.File, error) {
	return s.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY)
}

func (s *service) OpenFile(name tapr.PathName, flag int) (tapr.File, error) {
	s.data.mu.Lock()
	defer s.data.mu.Unlock()

	if (flag & os.O_TRUNC) == (flag & os.O_APPEND) {
		panic("O_TRUNC and O_APPEND are mutually exclusive")
	}

	b, ok := s.data.mu.blobs[name]
	if !ok {
		if flag&os.O_CREATE != 0 {
			b = make([]byte, 0)
			s.data.mu.blobs[name] = b

			return &file{
				name: name,
				buf:  bytes.NewBuffer(b),
			}, nil
		}

		return nil, &os.PathError{Op: "open", Path: string(name), Err: os.ErrNotExist}
	}

	// truncate if requested
	if flag&os.O_TRUNC != 0 {
		b = make([]byte, 0)
		s.data.mu.blobs[name] = b
	}

	buf := bytes.NewBuffer(b)

	if flag&os.O_APPEND != 0 {
		_ = buf.Next(len(b))
	}

	return &file{
		name: name,
		buf:  buf,
	}, nil
}

func (s *service) Open(name tapr.PathName) (tapr.File, error) {
	return s.OpenFile(name, 0)
}

func (s *service) Append(name tapr.PathName) (tapr.File, error) {
	return s.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY)
}

func (s *service) Mkdir(tapr.PathName) error {
	return nil
}

func (s *service) MkdirAll(tapr.PathName) error {
	return nil
}

func (s *service) Stat(path tapr.PathName) (os.FileInfo, error) {
	s.data.mu.Lock()
	defer s.data.mu.Unlock()

	b, ok := s.data.mu.blobs[path]
	if !ok {
		return nil, &os.PathError{Op: "stat", Path: string(path), Err: os.ErrNotExist}
	}

	return &fileInfo{
		file: &file{
			name: path,
			buf:  bytes.NewBuffer(b),
		},
	}, nil
}

type fileInfo struct {
	file *file
}

func (fi *fileInfo) Size() int64        { return int64(fi.file.buf.Len()) }
func (fi *fileInfo) IsDir() bool        { panic("not implemented for inprocess") }
func (fi *fileInfo) Mode() os.FileMode  { return os.ModePerm }
func (fi *fileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (fi *fileInfo) Sys() interface{}   { return fi.file }
func (fi *fileInfo) Name() string       { return string(fi.file.name) }
