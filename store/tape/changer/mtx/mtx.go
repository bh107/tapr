// Package mtx provides a changer.Changer that uses the mtx command to control
// a SCSI media changer.
package mtx

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"

	"tapr.space/errors"
	"tapr.space/store/tape"
	"tapr.space/store/tape/changer"
)

func init() {
	changer.Register("mtx", New)
}

var (
	hdrRegexp          = regexp.MustCompile(`\s*Storage Changer\s*(.*):(\d*) Drives, (\d*) Slots \( (\d*) Import/Export \)`)
	driveRegexp        = regexp.MustCompile(`Data Transfer Element (\d*):(.*)`)
	driveElementRegexp = regexp.MustCompile(`Full \(Storage Element (\d*) Loaded\):VolumeTag = (.*)`)
	slotRegexp         = regexp.MustCompile(`\s*Storage Element (\d*):(.*)`)
	mailSlotRegexp     = regexp.MustCompile(`\s*Storage Element (\d*) IMPORT/EXPORT:(.*)`)
	slotElementRegexp  = regexp.MustCompile(`Full :VolumeTag=(.*)`)
)

type changerImpl struct {
	path string
	prog string
}

var _ changer.Changer = (*changerImpl)(nil)

// New returns a new mtx tape implementation.
func New(opts map[string]interface{}) (changer.Changer, error) {
	const op = "changer/mtx.New"

	path, ok := opts["path"].(string)
	if !ok {
		return nil, errors.E(op, errors.Str("the path option must be specified"))
	}

	return &changerImpl{
		path: path,
		prog: "/usr/bin/mtx",
	}, nil
}

// do performs the given operation.
func (chgr *changerImpl) do(args ...string) ([]byte, error) {
	params := append([]string{"-f", chgr.path}, args...)

	return run(exec.Command(chgr.prog, params...))
}

func run(cmd *exec.Cmd) ([]byte, error) {
	var stderr bytes.Buffer

	if cmd.Stderr == nil {
		cmd.Stderr = &stderr
	}

	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return out, fmt.Errorf("%s: %s", exitError, stderr.String())
		}

		return out, err
	}

	return out, nil
}

// Load drive with the volume from slot.
func (chgr *changerImpl) Load(src, dst tape.Location) error {
	_, err := chgr.do(
		"load", strconv.Itoa(int(src.Addr)), strconv.Itoa(int(dst.Addr)),
	)

	return err
}

// Unload a volume from a drive and return it to a slot. If slotnum is zero,
// the volume will be returned to it's home slot.
func (chgr *changerImpl) Unload(src, dst tape.Location) error {
	_, err := chgr.do(
		"unload", strconv.Itoa(int(src.Addr)), strconv.Itoa(int(dst.Addr)),
	)

	return err
}

// Transfer moves a volume from one slot to another.
func (chgr *changerImpl) Transfer(src, dst tape.Location) error {
	_, err := chgr.do(
		"transfer", strconv.Itoa(int(src.Addr)), strconv.Itoa(int(dst.Addr)),
	)

	return err
}

// Status returns a Status structure with combined information about the status
// of the library.
func (chgr *changerImpl) Status() (map[tape.SlotCategory]tape.Slots, error) {
	status, err := chgr.do("status")
	if err != nil {
		return nil, err
	}

	return elements(status)
}

func elements(status []byte) (map[tape.SlotCategory]tape.Slots, error) {
	elements := map[tape.SlotCategory]tape.Slots{
		tape.TransferSlot:     make(tape.Slots, 0),
		tape.StorageSlot:      make(tape.Slots, 0),
		tape.ImportExportSlot: make(tape.Slots, 0),
	}

	scanner := bufio.NewScanner(bytes.NewReader(status))

	// skip header
	scanner.Scan()

	// scan elements
	var matches []string
	for scanner.Scan() {
		line := scanner.Text()

		// match data transfer elements
		matches = driveRegexp.FindStringSubmatch(line)
		if matches != nil {
			elemnum, err := strconv.Atoi(matches[1])
			if err != nil {
				return nil, err
			}

			s := tape.Slot{
				Location: tape.Location{
					Addr:     tape.Addr(elemnum),
					Category: tape.TransferSlot,
				},
			}

			if matches[2] != "Empty" {
				matches = driveElementRegexp.FindStringSubmatch(matches[2])
				if matches == nil {
					return nil, errors.Str("failed to parse transfer element")
				}

				home, err := strconv.Atoi(matches[1])
				if err != nil {
					return nil, err
				}

				s.Volume = &tape.Volume{
					Serial:   tape.Serial(matches[2]),
					Location: s.Location,
					Home: tape.Location{
						Addr: tape.Addr(home),
						// TODO(kbj): add category
					},
				}
			}

			elements[tape.TransferSlot] = append(elements[tape.TransferSlot], s)

			continue
		}

		// match storage elements
		matches = slotRegexp.FindStringSubmatch(line)
		if matches != nil {
			elemnum, err := strconv.Atoi(matches[1])
			if err != nil {
				return nil, err
			}

			s := tape.Slot{
				Location: tape.Location{
					Addr:     tape.Addr(elemnum),
					Category: tape.StorageSlot,
				},
			}

			if matches[2] != "Empty" {
				match := slotElementRegexp.FindStringSubmatch(matches[2])
				if match == nil {
					return nil, errors.Strf("failed to parse slot element: %v", matches[2])
				}

				s.Volume = &tape.Volume{
					Serial:   tape.Serial(match[1]),
					Location: s.Location,
				}
			}

			elements[tape.StorageSlot] = append(elements[tape.StorageSlot], s)

			continue
		}
		// match mailslot elements
		matches = mailSlotRegexp.FindStringSubmatch(line)
		if matches != nil {
			elemnum, err := strconv.Atoi(matches[1])
			if err != nil {
				return nil, err
			}

			s := tape.Slot{
				Location: tape.Location{
					Addr:     tape.Addr(elemnum),
					Category: tape.ImportExportSlot,
				},
			}

			if matches[2] != "Empty" {
				matches = slotElementRegexp.FindStringSubmatch(matches[2])
				if matches == nil {
					return nil, errors.Str("failed to parse slot element")
				}

				s.Volume = &tape.Volume{
					Serial:   tape.Serial(matches[1]),
					Location: s.Location,
				}
			}

			elements[tape.ImportExportSlot] = append(elements[tape.ImportExportSlot], s)

			continue
		}

		return nil, errors.Strf("failed to parse slot")
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return elements, nil
}

func params(status []byte) (map[string]int, error) {
	params := make(map[string]int)

	scanner := bufio.NewScanner(bytes.NewReader(status))

	var err error
	if scanner.Scan() {
		line := scanner.Text()

		matches := hdrRegexp.FindStringSubmatch(line)
		if matches == nil {
			return nil, errors.Strf("failed to match mtx status header")
		}

		params["maxDrives"], err = strconv.Atoi(matches[2])
		if err != nil {
			return nil, err
		}

		params["numSlots"], err = strconv.Atoi(matches[3])
		if err != nil {
			return nil, err
		}

		params["numMailSlots"], err = strconv.Atoi(matches[4])
		if err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return params, nil
}
