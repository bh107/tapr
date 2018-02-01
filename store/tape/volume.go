package tape

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"tapr.space/bitmask"
)

// A Serial is the volume serial number (VOLSER) of a tape.
type Serial string

// VolumeCategory represents the volume category.
type VolumeCategory int

// Known volume categories.
const (
	UnknownVolume VolumeCategory = iota
	Allocating
	Scratch
	Filling
	Full
	Missing
	Damaged
	Cleaning
)

const (
	// StatusTransfering denotes that the volume is currently being transfered by the
	// media changer.
	StatusTransfering uint32 = 1 << iota

	// StatusMounted is when the volume is mounted.
	StatusMounted

	// StatusNeedsCleaning tells us that the volume needs cleaning.
	StatusNeedsCleaning
)

// FormatVolumeFlags formats the flags for human consumption.
func FormatVolumeFlags(f uint32) string {
	var out []string
	if bitmask.IsSet(f, StatusTransfering) {
		out = append(out, "transfering")
	}

	if bitmask.IsSet(f, StatusMounted) {
		out = append(out, "mounted")
	}

	if bitmask.IsSet(f, StatusNeedsCleaning) {
		out = append(out, "needs-cleaning")
	}

	return strings.Join(out, ",")
}

// A Volume is an usable volume.
type Volume struct {
	// Serial is the Volume Serial (VOLSER).
	Serial Serial

	// Location is the current location in the store.
	Location Location

	// Home is the home location in the store.
	Home Location

	// Category tracks the volume state.
	Category VolumeCategory

	// Flags are contains temporary info on the volume.
	Flags uint32
}

func (v *Volume) String() string {
	return fmt.Sprintf("%s/%s,%s)", v.Serial, v.Category, FormatVolumeFlags(v.Flags))
}

// String implements fmt.Stringer.
func (cat VolumeCategory) String() string {
	switch cat {
	case UnknownVolume:
		return "unknown"
	case Allocating:
		return "allocating"
	case Scratch:
		return "scratch"
	case Filling:
		return "filling"
	case Full:
		return "full"
	case Missing:
		return "missing"
	case Damaged:
		return "damaged"
	case Cleaning:
		return "cleaning"
	}

	panic("unknown volume category")
}

// ToVolumeCategory returns the VolumeStatus corresponding to the given string.
func ToVolumeCategory(str string) VolumeCategory {
	switch str {
	case "unknown":
		return UnknownVolume
	case "allocating":
		return Allocating
	case "scratch":
		return Scratch
	case "filling":
		return Filling
	case "full":
		return Full
	case "missing":
		return Missing
	case "damaged":
		return Damaged
	case "cleaning":
		return Cleaning
	default:
		panic("unknown volume category")
	}
}

// Value implements driver.Valuer.
func (cat VolumeCategory) Value() (driver.Value, error) {
	return cat.String(), nil
}

// Scan implements sql.Scanner.
func (cat *VolumeCategory) Scan(src interface{}) error {
	*cat = ToVolumeCategory(string(src.([]byte)))

	return nil
}
