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

package config // import "tapr.space/config"

import (
	"io"
	"io/ioutil"
	"os"
	osuser "os/user"

	yaml "gopkg.in/yaml.v2"

	"tapr.space"
	"tapr.space/errors"
)

// A StoreConfigurator is a function that given a raw config YAML value
// produces a implementation specific store config.
type StoreConfigurator func(cfg *StoreConfig, unmarshal func(interface{}) error) error

var registration = make(map[string]StoreConfigurator)

// Register registers the StoreConfigurator with the config system.
func Register(name string, fn StoreConfigurator) error {
	const op = "config.Register"
	if _, exists := registration[name]; exists {
		return errors.E(op, errors.Exist)
	}

	registration[name] = fn

	return nil
}

// StoreConfig is a partial store configuration.
type StoreConfig struct {
	Backend string

	// Embedded holds an embedded implementation-specific configuration.
	Embedded interface{}
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *StoreConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	m := make(map[string]interface{})
	if err := unmarshal(&m); err != nil {
		return err
	}

	name, ok := m["backend"].(string)
	if !ok {
		return errors.E(errors.Invalid, errors.Strf("backend must be a string"))
	}

	c.Backend = name

	if fn, exists := registration[name]; exists {
		return fn(c, unmarshal)
	}

	return errors.E(errors.Invalid, errors.Strf("no configurator for backend type '%v'", name))
}

// ServerConfig is the main server configuration.
type ServerConfig struct {
	Stores map[string]StoreConfig `yaml:"stores"`
}

// InitServerConfig initializes a server configuration from an io.Reader.
func InitServerConfig(r io.Reader) (*ServerConfig, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var cfg ServerConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// base implements tapr.Config, returning default values for all operations.
type base struct{}

func (base) Target() string      { return "default" }
func (base) Value(string) string { return "" }

// New returns a config with all fields set as defaults.
func New() tapr.Config {
	return base{}
}

// InitConfig initializes a client config from an io.Reader.
func InitConfig(r io.Reader) (tapr.Config, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	cfg := New()
	_cfg := make(map[string]string)
	if err := yaml.Unmarshal(b, _cfg); err != nil {
		return nil, err
	}

	SetTarget(cfg, _cfg["target"])

	return cfg, nil
}

type cfgTarget struct {
	tapr.Config
	target string
}

func (cfg cfgTarget) Target() string {
	return cfg.target
}

// SetTarget returns a config derived from the given config with the
// target store set.
func SetTarget(cfg tapr.Config, target string) tapr.Config {
	return cfgTarget{
		Config: cfg,
		target: target,
	}
}

// Homedir returns the home directory of the OS' logged-in user.
func Homedir() (string, error) {
	u, err := osuser.Current()

	// user.Current may return an error, but we should only handle it if it
	// returns a nil user. This is because os/user is wonky without cgo,
	// but it should work well enough for our purposes.
	if u == nil {
		e := errors.Str("lookup of current user failed")

		if err != nil {
			e = errors.Strf("%v: %v", e, err)
		}

		return "", e
	}

	h := u.HomeDir
	if h == "" {
		return "", errors.E(errors.NotExist, errors.Str("user home directory not found"))
	}

	if err := isDir(h); err != nil {
		return "", err
	}
	return h, nil
}

func isDir(p string) error {
	fi, err := os.Stat(p)
	if err != nil {
		return errors.E(errors.IO, err)
	}
	if !fi.IsDir() {
		return errors.E(errors.NotDir, errors.Str(p))
	}
	return nil
}
