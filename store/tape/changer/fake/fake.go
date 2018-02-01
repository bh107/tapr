// Package fake provides a fake tape.Changer.
package fake

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"tapr.space/bitmask"
	"tapr.space/errors"
	"tapr.space/log"
	"tapr.space/sim"
	"tapr.space/store/tape"
	"tapr.space/store/tape/changer"
)

func init() {
	changer.Register("fake", New)
}

type changerImpl struct {
	mu sync.Mutex

	slots map[tape.SlotCategory]tape.Slots
}

var _ changer.Changer = (*changerImpl)(nil)

// New returns a new fake tape implementation.
func New(opts map[string]interface{}) (changer.Changer, error) {
	const op = "tape/fake.New"

	var vols []tape.Volume

	requiredOpts := []string{
		"transfer", "storage", "ix", "volumes",
	}

	sopts := make(map[string]int)

	for _, opt := range requiredOpts {
		if _, ok := opts[opt]; !ok {
			return nil, errors.E(op, errors.Strf("the %s option must be specified", opt))
		}

		iopt, err := strconv.Atoi(opts[opt].(string))
		if err != nil {
			return nil, err
		}

		sopts[opt] = iopt
	}

	if v, ok := opts["vols"]; ok {
		vols = v.([]tape.Volume)
	}

	chgr := changerImpl{
		slots: make(map[tape.SlotCategory]tape.Slots),
	}

	slots := make(tape.Slots, sopts["transfer"])
	for i := range slots {
		slots[i] = tape.Slot{
			Location: tape.Location{
				Addr:     tape.Addr(i),
				Category: tape.TransferSlot,
			},
		}
	}
	chgr.slots[tape.TransferSlot] = slots

	slots = make(tape.Slots, 0)
	// storage slots are usually numbered from 1, so insert an invalid slot.
	slots = append(slots, tape.Slot{
		Location: tape.Location{
			Addr:     tape.Addr(0),
			Category: tape.InvalidSlot,
		},
	})

	for i := 0; i < sopts["storage"]; i++ {
		slot := tape.Slot{
			Location: tape.Location{
				Addr:     tape.Addr(i + 1),
				Category: tape.StorageSlot,
			},
		}

		if vols == nil {
			if i < sopts["volumes"] {
				slot.Volume = &tape.Volume{
					Serial:   tape.Serial(fmt.Sprintf("%c%05dL7", 'A', i)),
					Location: slot.Location,
				}
			}
		}

		slots = append(slots, slot)
	}

	slots[len(slots)-1].Volume = &tape.Volume{
		Serial:   tape.Serial("CLN000L1"),
		Location: slots[len(slots)-1].Location,
	}

	chgr.slots[tape.StorageSlot] = slots

	slots = make(tape.Slots, 0)
	for i := 0; i < sopts["ix"]; i++ {
		slots = append(slots, tape.Slot{
			Location: tape.Location{
				Addr:     tape.Addr(sopts["storage"] + i + 1),
				Category: tape.ImportExportSlot,
			},
		})
	}
	chgr.slots[tape.ImportExportSlot] = slots

	for i, v := range vols {
		chgr.slots[v.Location.Category][v.Location.Addr].Volume = &vols[i]
	}

	return &chgr, nil
}

func (chgr *changerImpl) Status() (map[tape.SlotCategory]tape.Slots, error) {
	chgr.mu.Lock()
	defer chgr.mu.Unlock()

	sim.Maybe(func(state sim.State) {
		// simulate status
		state.Simulate(&sim.NormalDistributedNoise{
			Mean:   1 * time.Second,
			Stddev: 100 * time.Millisecond,
		})
	})

	return chgr.slots, nil
}

func (chgr *changerImpl) Unload(src, dst tape.Location) error {
	chgr.mu.Lock()
	defer chgr.mu.Unlock()

	const op = "tape/fake.Unload"

	srcSlot := &chgr.slots[src.Category][src.Addr]
	dstSlot := &chgr.slots[dst.Category][dst.Addr]

	sim.Maybe(func(state sim.State) {
		v := srcSlot.Volume

		bitmask.Set(&v.Flags, tape.StatusTransfering)

		srcSlot.Volume = nil

		// simulate transfer
		state.Simulate(&sim.NormalDistributedNoise{
			Mean:   1 * time.Second,
			Stddev: 100 * time.Millisecond,
		})

		// update locations
		v.Location = dst
		v.Home = tape.Location{Category: tape.InvalidSlot}

		dstSlot.Volume = v

		bitmask.Clear(&v.Flags, tape.StatusTransfering)
	})

	return nil
}

func (chgr *changerImpl) Load(src, dst tape.Location) error {
	chgr.mu.Lock()
	defer chgr.mu.Unlock()

	const op = "tape/fake.Load"

	srcSlot := &chgr.slots[src.Category][src.Addr]
	dstSlot := &chgr.slots[dst.Category][dst.Addr]

	log.Debug.Printf("%s: loading from %v to %v", op, srcSlot, dstSlot)

	sim.Maybe(func(state sim.State) {
		v := srcSlot.Volume

		bitmask.Set(&v.Flags, tape.StatusTransfering)

		srcSlot.Volume = nil

		// simulate transfer
		state.Simulate(&sim.NormalDistributedNoise{
			Mean:   1 * time.Second,
			Stddev: 100 * time.Millisecond,
		})

		// update locations
		v.Location = dst
		v.Home = src

		dstSlot.Volume = v

		bitmask.Clear(&v.Flags, tape.StatusTransfering)
	})

	return nil
}

func (chgr *changerImpl) Transfer(src, dst tape.Location) error {
	chgr.mu.Lock()
	defer chgr.mu.Unlock()

	const op = "tape/fake.Transfer"

	srcSlot := &chgr.slots[src.Category][src.Addr]
	dstSlot := &chgr.slots[dst.Category][dst.Addr]

	sim.Maybe(func(state sim.State) {
		v := srcSlot.Volume

		bitmask.Set(&v.Flags, tape.StatusTransfering)

		srcSlot.Volume = nil

		// simulate transfer
		state.Simulate(&sim.NormalDistributedNoise{
			Mean:   3 * time.Second,
			Stddev: 1 * time.Second,
		})

		v.Location = dst

		dstSlot.Volume = v

		bitmask.Clear(&v.Flags, tape.StatusTransfering)
	})

	return nil
}
