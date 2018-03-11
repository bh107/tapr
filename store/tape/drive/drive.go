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
	Serial  tape.Serial

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

	loaded, serial, err := invdb.Loaded(drv.loc)
	if err != nil {
		return err
	}

	if !loaded {
		log.Debug.Printf("%s: drive is empty, allocating", op)
		// get a volume from the inventory if we do not already have a
		// volume mounted
		serial, err = invdb.Alloc()
		if err != nil {
			return err
		}

		log.Debug.Printf("%s: loading %v into %v", op, serial, drv.loc)

		if err := invdb.Load(serial, drv.loc, chgr); err != nil {
			return err
		}
	}

	vol, err := invdb.Info(serial)
	if err != nil {
		return err
	}

	// format the volume if necessary
	formatted, stg, err := fmtr.Format(drv.devpath, vol)
	if err != nil {
		return err
	}

	if formatted {
		vol.Category = tape.Filling
		if err := invdb.Update(vol); err != nil {
			return err
		}
	}

	// mount if needed
	if mounter, ok := stg.(format.Mounter); ok {
		if err := mounter.Mount(); err != nil {
			return err
		}
	}

	drv.Serial = serial
	drv.Storage = stg

	return nil
}
