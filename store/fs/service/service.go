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

// Package service implements a simple store.Store using an existing file system.
package service

import (
	"os"

	"tapr.space/config"
	"tapr.space/log"
	"tapr.space/storage/fsdir"
	"tapr.space/store"
)

func init() {
	config.Register("store/fs", configurator)
	store.Register("store/fs", New)
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
