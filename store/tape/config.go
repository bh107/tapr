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
