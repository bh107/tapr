// Package scsi provides a SCSI tape drive implementation.
package scsi // import "hpt.space/tapr/store/tape/drive/scsi"

import (
	"hpt.space/tapr/errors"
	"hpt.space/tapr/store/tape/changer"
	"hpt.space/tapr/store/tape/drive"
	"hpt.space/tapr/store/tape/inv"
)

func init() {
	drive.Register("scsi", New)
}

type impl struct {
	path string
}

var _ drive.Drive = (*impl)(nil)

// New returns a new scsi tape drive implementation.
func New(opts map[string]interface{}) (drive.Drive, error) {
	const op = "drive/scsi.New"

	path, ok := opts["path"].(string)
	if !ok {
		return nil, errors.E(op, errors.Str("the path option must be specified"))
	}

	return &impl{
		path: path,
	}, nil
}

func (drv *impl) Setup(inv inv.Inventory, chgr changer.Changer) {
	panic("not implemented")
}
