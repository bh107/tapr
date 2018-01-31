package inv // import "hpt.space/tapr/store/tape/inv"

import (
	"hpt.space/tapr"
	"hpt.space/tapr/errors"
	"hpt.space/tapr/store/tape"
	"hpt.space/tapr/store/tape/changer"
)

// InventoryConstructor is a function that creates an Inventory.
type InventoryConstructor func(map[string]string) (Inventory, error)

var registration = make(map[string]InventoryConstructor)

// Register registers a new inv.Inventory implementation.
func Register(name string, fn InventoryConstructor) error {
	const op = "inv.Register"
	if _, exists := registration[name]; exists {
		return errors.E(op, errors.Exist)
	}

	registration[name] = fn

	return nil
}

// Create creates a new Inventory using the named implementation.
func Create(name string, cfg map[string]string) (Inventory, error) {
	const op = "store.Create"

	fn, found := registration[name]
	if !found {
		return nil, errors.E(op, errors.Invalid, errors.Strf("unknown inventory backend type: %v", name))
	}

	return fn(cfg)
}

// An Inventory tracks volumes in a tape store. An inventory MUST be safe for
// concurrent use.
type Inventory interface {
	// Load transfers a volume to a drive in the context of
	// the given tape.
	Load(tape.Serial, tape.Location, changer.Changer) error

	// Unload transfers a volume from a drive in the context of
	// the given tape.
	Unload(tape.Serial, tape.Location, changer.Changer) error

	// Transfer transfers a volume to a location in the context of
	// the given tape.
	Transfer(tape.Serial, tape.Location, changer.Changer) error

	// Audit performs an inventory audit by inspecting the physical state of the
	// backing tape and makes sure the inventory reflects that state.
	Audit(changer.Changer) error

	// Alloc allocates a filling (or scratch) volume from the inventory
	// and mounts it into the device at the given destination.
	Alloc() (tape.Volume, error)

	// Status returns a list of known volumes.
	Volumes() ([]tape.Volume, error)

	// Reset resets the inventory database.
	Reset() error

	// Lookup looks up the given path name and returns a list of volumes that
	// contains the file/directory.
	Lookup(tapr.PathName) (tape.Volume, error)

	// Create creates a new node in the directory tree located on the volume
	// associated with the given volume serial.
	Create(path tapr.PathName, serial string) error
}
