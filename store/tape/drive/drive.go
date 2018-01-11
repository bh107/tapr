package drive // import "hpt.space/tapr/store/tape/drive"

import (
	"hpt.space/tapr/errors"
	"hpt.space/tapr/log"
	"hpt.space/tapr/store/tape"
	"hpt.space/tapr/store/tape/changer"
	"hpt.space/tapr/store/tape/inv"
)

// Launch launches a new iodev process, managing a mounted volume.
func Launch(loc tape.Location, vol *tape.Volume, invdb inv.Inventory, chgr changer.Changer) error {
	const op = "iodev.Create"

	if vol == nil {
		// get a volume from the inventory if we do not already have a
		// volume mounted
		v, err := invdb.Alloc()
		if err != nil {
			return err
		}

		log.Debug.Printf("%s: loading %v into %v", op, v, loc)

		if err := invdb.Load(v.Serial, loc, chgr); err != nil {
			return err
		}

		log.Debug.Printf("%s: loaded %v into %v", op, v, loc)

		return nil
	}

	log.Debug.Printf("%s: volume %v already loaded in %v", op, vol, loc)

	return nil
}

// Constructor is a function that creates a Drive.
type Constructor func(map[string]interface{}) (Drive, error)

var registration = make(map[string]Constructor)

// Register registers a new drive.Drive implementation.
func Register(name string, fn Constructor) error {
	const op = "drive.Register"
	if _, exists := registration[name]; exists {
		return errors.E(op, errors.Exist)
	}

	registration[name] = fn

	return nil
}

// Create creates a new Drive using the named implementation.
func Create(name string, cfg map[string]interface{}) (Drive, error) {
	const op = "drive.Create"

	fn, found := registration[name]
	if !found {
		return nil, errors.E(op, errors.Invalid, errors.Strf("unknown drive backend type: %v", name))
	}

	return fn(cfg)
}

// A Drive is a tape drive.
type Drive interface {
}
