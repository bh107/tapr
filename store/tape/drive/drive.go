package drive // import "tapr.space/store/tape/drive"

import (
	"fmt"
	"os"

	"tapr.space/flags"
	"tapr.space/format"
	"tapr.space/log"
	"tapr.space/storage"
	"tapr.space/store/tape"
	"tapr.space/store/tape/changer"
	"tapr.space/store/tape/inv"
)

// Drive represents a tape drive.
type Drive struct {
	devpath string
	serial  string

	name string
	loc  tape.Location

	storage.Storage
}

// New returns a new fake tape drive implementation.
func New(name string, cfg tape.DriveConfig) (*Drive, error) {
	op := fmt.Sprintf("drive/Drive.New[%s (slot %d) (path %s)]", name, cfg.Slot, cfg.Path)

	if flags.EmulateDevices {
		if err := os.MkdirAll(cfg.Path, os.ModePerm); err != nil {
			return nil, err
		}
	}

	loc := tape.Location{
		Addr:     tape.Addr(cfg.Slot),
		Category: tape.TransferSlot,
	}

	log.Debug.Printf("%s: created", op)

	drv := &Drive{
		devpath: cfg.Path,
		name:    name,
		loc:     loc,
	}

	return drv, nil
}

// Start the drive.
func (drv *Drive) Start(invdb inv.Inventory, chgr changer.Changer, fmtr format.Formatter) error {
	op := fmt.Sprintf("drive/fake.Setup[%s (slot %d) (path %s)]", drv.name, drv.loc.Addr, drv.devpath)

	loaded, vol, err := invdb.Loaded(drv.loc)
	if err != nil {
		return err
	}

	if !loaded {
		log.Debug.Printf("%s: drive is empty, allocating", op)
		// get a volume from the inventory if we do not already have a
		// volume mounted
		vol, err = invdb.Alloc()
		if err != nil {
			return err
		}

		log.Debug.Printf("%s: loading %v into %v", op, vol.Serial, drv.loc)

		if err := invdb.Load(vol.Serial, drv.loc, chgr); err != nil {
			return err
		}
	}

	// format the volume
	stg, err := fmtr.Format(drv.devpath, vol.Serial)
	if err != nil {
		return err
	}

	// mount if needed
	if mounter, ok := stg.(format.Mounter); ok {
		if err := mounter.Mount(); err != nil {
			return err
		}
	}

	drv.Storage = stg

	return nil
}
