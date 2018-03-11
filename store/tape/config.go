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

package tape

import "tapr.space/config"

func init() {
	config.Register("store/tape", configurator)
}

func configurator(c *config.StoreConfig, unmarshal func(interface{}) error) error {
	var cfg Config
	if err := unmarshal(&cfg); err != nil {
		return err
	}

	c.Embedded = cfg

	return nil
}

// Config is the tape config.
type Config struct {
	// The storage format ("ltfs", "bltfs", "raw")
	Format struct {
		Driver  string
		Options map[string]string
	}

	// CleaningPrefix is the prefix that identifies cleaning cartridges.
	CleaningPrefix string `yaml:"cleaning-prefix"`

	// Inventory is the inventory database configuration.
	Inventory struct {
		Driver  string
		Options map[string]string
	}

	// Changers contains configuration for the media changers.
	Changers map[string]ChangerConfig

	// Drives contains configuration for the drives.
	Drives struct {
		Format FormatConfig

		Read  map[string]DriveConfig
		Write map[string]DriveConfig
	}
}

// ChangerConfig holds the configuration for a changer.
type ChangerConfig struct {
	Driver  string
	Options map[string]string
}

// DriveConfig holds configuration for drives.
type DriveConfig struct {
	Slot int
	Path string
}

type FormatConfig struct {
	Backend string
	Options map[string]string
}
