package tape

import (
	"database/sql/driver"
	"fmt"
)

// Addr is an element/slot address in a store.
type Addr int

// Location uniquely identifies a location within a store.
type Location struct {
	Addr     Addr
	Category SlotCategory
}

type SlotMap map[Location]Volume

// SlotCategory is the type of slot (unknown, invalid, data transfer, storage
// or import/export).
type SlotCategory int

const (
	// UnknownSlot represents slots which purpose is unknown.
	UnknownSlot SlotCategory = iota

	// InvalidSlot represents slots that are invalidly addressed.
	InvalidSlot

	// TransferSlot represents data transfer slots, usually inhabited
	// by a tape drive.
	TransferSlot

	// StorageSlot represents storage slots.
	StorageSlot

	// ImportExportSlot represents so-called mailbox slots for bulk import/export
	// of volumes to and from the silo.
	ImportExportSlot
)

// SlotCategories is a list of all valid slot types.
var SlotCategories = []SlotCategory{TransferSlot, StorageSlot, ImportExportSlot}

// Value implements the driver.Valuer interface.
func (loc *Location) Value() (driver.Value, error) {
	return fmt.Sprintf("(%d,%s)", loc.Addr, loc.Category), nil
}

// Scan implements sql.Scanner.
func (loc *Location) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	v := src.([]byte)

	var addr int
	var t string

	// remove parantheses
	str := string(v[1 : len(v)-1])

	n, err := fmt.Sscanf(str, "%d,%s", &addr, &t)
	if err != nil {
		return err
	}

	if n != 2 {
		return err
	}

	loc.Addr = Addr(addr)
	loc.Category = ToSlotCategory(t)

	return nil
}

// String implements fmt.Stringer.
func (cat SlotCategory) String() string {
	switch cat {
	case UnknownSlot:
		return "unknown"
	case InvalidSlot:
		return "invalid"
	case TransferSlot:
		return "transfer"
	case StorageSlot:
		return "storage"
	case ImportExportSlot:
		return "ix"
	default:
		panic("unknown slot category")
	}
}

// ToSlotCategory returns the SlotCategory corresponding to the given string.
func ToSlotCategory(str string) SlotCategory {
	switch str {
	case "unknown":
		return UnknownSlot
	case "invalid":
		return InvalidSlot
	case "transfer":
		return TransferSlot
	case "storage":
		return StorageSlot
	case "ix":
		return ImportExportSlot
	default:
		panic("unknown slot category")
	}
}

// Value implements driver.Valuer.
func (cat SlotCategory) Value() (driver.Value, error) {
	return cat.String(), nil
}

// Slots is a list of slots.
type Slots []Slot

// Slot is a slot in a store, represented by its element address.
type Slot struct {
	Location

	// Volume returns the volume currently in this slot (if any).
	Volume *Volume
}

func (s *Slot) String() string {
	return fmt.Sprintf("(%d,%s)", s.Addr, s.Category)
}
