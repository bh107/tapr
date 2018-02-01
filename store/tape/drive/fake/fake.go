// Package fake provides a fake tape drive implementation.
package fake // import "tapr.space/store/tape/drive/fake"

import (
	"fmt"

	"tapr.space/log"
	"tapr.space/store/tape"
	"tapr.space/store/tape/changer"
	"tapr.space/store/tape/drive"
	"tapr.space/store/tape/inv"
)

func init() {
	drive.Register("fake", New)
}

type impl struct {
	name string
	loc  tape.Location
	vol  *tape.Volume

	stop chan struct{}
}

var _ drive.Drive = (*impl)(nil)

// New returns a new fake tape drive implementation.
func New(name string, cfg tape.DriveConfig) (drive.Drive, error) {
	op := fmt.Sprintf("drive/fake.New[%s (slot %d)]", name, cfg.Slot)

	loc := tape.Location{
		Addr:     tape.Addr(cfg.Slot),
		Category: tape.TransferSlot,
	}

	log.Debug.Printf("%s: created", op)

	drv := &impl{
		name: name,
		loc:  loc,
		stop: make(chan struct{}),
	}

	return drv, nil
}

func (drv *impl) Setup(invdb inv.Inventory, chgr changer.Changer) {
	op := fmt.Sprintf("drive/fake.Setup[%s (slot %d)]", drv.name, drv.loc.Addr)

	loaded, err := invdb.Loaded(drv.loc)
	if err != nil {
		log.Fatal(err)
	}

	if !loaded {
		log.Debug.Printf("%s: drive is empty, allocating", op)
		// get a volume from the inventory if we do not already have a
		// volume mounted
		v, err := invdb.Alloc()
		if err != nil {
			log.Fatal(err)
		}

		log.Debug.Printf("%s: loading %v into %v", op, v, drv.loc)

		if err := invdb.Load(v.Serial, drv.loc, chgr); err != nil {
			log.Fatal(err)
		}
	}
}
