package proto // import "tapr.space/store/tape/proto"

import "tapr.space/store/tape"

// To regenerate the protocol buffer output for this package, run
//      go generate

//go:generate protoc tape.proto --go_out=.

// VolumeProto converts a tapr.Volume to a proto.Volume.
func VolumeProto(v tape.Volume) *Volume {
	return &Volume{
		Serial:   string(v.Serial),
		Location: LocationProto(v.Location),
		Home:     LocationProto(v.Home),
		Category: VolumeCategoryProto(v.Category),
		Flags:    uint32(v.Flags),
	}
}

// VolumeProtos converts a slice of tapr.Volume to a slice of proto.Volume.
func VolumeProtos(vols []tape.Volume) []*Volume {
	if len(vols) == 0 {
		return nil
	}

	pbs := make([]*Volume, len(vols))
	for i := range pbs {
		pbs[i] = VolumeProto(vols[i])
	}

	return pbs
}

// TaprVolume converts a proto.Volume to a tapr.Volume.
func TaprVolume(pb *Volume) tape.Volume {
	if pb == nil {
		return tape.Volume{}
	}

	return tape.Volume{
		Serial:   tape.Serial(pb.Serial),
		Location: TaprLocation(pb.Location),
		Home:     TaprLocation(pb.Home),
		Category: TaprVolumeCategory(pb.Category),
		Flags:    pb.Flags,
	}
}

// TaprVolumes converts a slice of proto.Volume to a slice of tapr.Volume.
func TaprVolumes(pbs []*Volume) []tape.Volume {
	if len(pbs) == 0 {
		return nil
	}

	vols := make([]tape.Volume, len(pbs))
	for i := range vols {
		vols[i] = TaprVolume(pbs[i])
	}

	return vols
}

// VolumeCategoryProto converts a tapr.VolumeCategory to a proto.Volume_Category.
func VolumeCategoryProto(s tape.VolumeCategory) Volume_Category {
	switch s {
	case tape.UnknownVolume:
		return Volume_UNKNOWN
	case tape.Allocating:
		return Volume_ALLOCATING
	case tape.Filling:
		return Volume_FILLING
	case tape.Scratch:
		return Volume_SCRATCH
	case tape.Full:
		return Volume_FULL
	case tape.Missing:
		return Volume_MISSING
	case tape.Damaged:
		return Volume_DAMAGED
	case tape.Cleaning:
		return Volume_CLEANING
	default:
		panic("unknown volume category")
	}
}

// TaprVolumeCategory converts a proto.Volume_Category to a tapr.VolumeCategory.
func TaprVolumeCategory(pb Volume_Category) tape.VolumeCategory {
	switch pb {
	case Volume_UNKNOWN:
		return tape.UnknownVolume
	case Volume_ALLOCATING:
		return tape.Allocating
	case Volume_FILLING:
		return tape.Filling
	case Volume_SCRATCH:
		return tape.Scratch
	case Volume_FULL:
		return tape.Full
	case Volume_MISSING:
		return tape.Missing
	case Volume_DAMAGED:
		return tape.Damaged
	case Volume_CLEANING:
		return tape.Cleaning
	default:
		panic("unknown volume category")
	}
}

// LocationProto converts a tapr.Location to a proto.Location.
func LocationProto(loc tape.Location) *Location {
	return &Location{
		Addr:     int64(loc.Addr),
		Category: loc.Category.String(),
	}
}

// TaprLocation converts a proto.Location to a tapr.Location.
func TaprLocation(pb *Location) tape.Location {
	return tape.Location{
		Addr:     tape.Addr(pb.Addr),
		Category: tape.ToSlotCategory(pb.Category),
	}
}
