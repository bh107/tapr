package tape

import "hpt.space/tapr/config"

func init() {
	config.Register("tape", configurator)
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
	Format string `yaml:"format"`

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
	Slot    int
	Driver  string
	Options map[string]string
}
