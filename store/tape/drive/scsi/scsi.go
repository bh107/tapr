// Package scsi provides a SCSI tape drive implementation.
package scsi // import "tapr.space/store/tape/drive/scsi"

import (
	"tapr.space/errors"
	"tapr.space/store/tape"
	"tapr.space/store/tape/changer"
	"tapr.space/store/tape/drive"
	"tapr.space/store/tape/inv"
)

func init() {
	drive.Register("scsi", New)
}

type impl struct {
	path string
}

var _ drive.Drive = (*impl)(nil)

// New returns a new scsi tape drive implementation.
func New(name string, cfg tape.DriveConfig) (drive.Drive, error) {
	const op = "drive/scsi.New"

	path, ok := cfg.Options["path"]
	if !ok {
		return nil, errors.E(op, errors.Str("the path option must be specified"))
	}

	return &impl{
		path: path,
	}, nil
}

func (drv *impl) Setup(invdb inv.Inventory, chgr changer.Changer) {
	panic("not implemented")
}
