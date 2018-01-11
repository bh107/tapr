// Package service implements a simple store.Store using an existing file system.
package service

import (
	"os"

	"hpt.space/tapr/config"
	"hpt.space/tapr/log"
	"hpt.space/tapr/storage/fsdir"
	"hpt.space/tapr/store"
)

func init() {
	config.Register("fs", configurator)
	store.Register("fs", New)
}

func configurator(c *config.StoreConfig, unmarshal func(interface{}) error) error {
	var cfg Config
	if err := unmarshal(&cfg); err != nil {
		return err
	}

	c.Embedded = cfg

	return nil
}

// Config holds the configuration for a file system backed fs.Store
// implementation.
type Config struct {
	Root string `yaml:"root"`
}

type service struct {
	name string

	*fsdir.Storage
}

var _ store.Store = (*service)(nil)

// New creates a new store.Store service.
func New(name string, _cfg config.StoreConfig) (store.Store, error) {
	op := "store/fs/service.New[" + name + "]"
	cfg := _cfg.Embedded.(Config)

	log.Debug.Printf("%s: creating store", op)

	root := cfg.Root

	if err := os.MkdirAll(root, os.ModePerm); err != nil {
		return nil, err
	}

	svc := &service{
		name:    name,
		Storage: fsdir.New(root),
	}

	return svc, nil
}

func (s *service) String() string {
	return s.name
}
