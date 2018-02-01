// +build linux

// Package ltfs provides simple functions to work with LTFS volumes
// through the reference LTFS implementation tools.
package ltfs

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"tapr.space/errors"
	"tapr.space/flags"
	"tapr.space/format"
	"tapr.space/log"
	"tapr.space/storage"
	"tapr.space/storage/fsdir"
	"tapr.space/store/tape"
)

const (
	ltfsCommand       = "/usr/local/bin/ltfs"
	fusermountCommand = "/usr/bin/fusermount"
	mkltfsCommand     = "/usr/local/bin/mkltfs"
)

func init() {
	format.Register("ltfs", New)
}

type fmtr struct {
	mountdir string
}

type impl struct {
	proc *os.Process

	devpath   string
	mountpath string

	// embed an fsdir storage implementation
	*fsdir.Storage
}

var _ storage.Storage = (*impl)(nil)

// New create a new LTFS format.
func New(cfg tape.FormatConfig) (format.Formatter, error) {
	mountdir, ok := cfg.Options["mountdir"]
	if !ok {
		return nil, errors.E(errors.Invalid, "missing moundir option")
	}

	if err := os.MkdirAll(mountdir, os.ModePerm); err != nil {
		return nil, err
	}

	return &fmtr{
		mountdir: mountdir,
	}, nil
}

// Format formats the volume.
func (f *fmtr) Format(devpath string, vol tape.Volume) (formatted bool, stg storage.Storage, err error) {
	fi, err := os.Stat(devpath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil, errors.E(errors.Invalid, errors.Strf("%s does not exist", devpath))
		}

		return false, nil, err
	}

	if flags.EmulateDevices {
		if !fi.IsDir() {
			return false, nil, errors.E(errors.Invalid, errors.Strf("%s is not a directory", devpath))
		}
	}

	if vol.Category == tape.Allocated {
		opts := []string{
			fmt.Sprintf("--device=%s", devpath),
			fmt.Sprintf("--tape-serial=%s", vol.Serial[:6]),
		}

		if flags.EmulateDevices {
			opts = append(opts, fmt.Sprintf("--backend=file"))
		}

		cmd := exec.Command(mkltfsCommand, opts...)

		log.Debug.Printf("running: %s (args: %v)", cmd.Path, cmd.Args[1:])

		_, err = execCmd(cmd)
		if err != nil {
			return false, nil, errors.E(err, "failed to format volume")
		}

		formatted = true
	}

	return formatted, &impl{
		devpath:   devpath,
		mountpath: filepath.Join(f.mountdir, string(vol.Serial)),
	}, nil
}

// Mount mounts the LTFS formatted volume loaded in the device pointed
// to by devpath into the path directory. The function returns
// an *os.Process refering to the background ltfs process responsible
// for the mounted file system.
func (f *impl) Mount() error {
	if err := os.MkdirAll(f.mountpath, os.ModePerm); err != nil {
		return err
	}

	opts := []string{
		f.mountpath,

		"-o", fmt.Sprintf("devname=%s", f.devpath),
		"-o", fmt.Sprintf("sync_type=%s", "unmount"),
	}

	if flags.EmulateDevices {
		opts = append(opts, "-o", "tape_backend=file")
	}

	cmd := exec.Command(ltfsCommand, opts...)

	log.Debug.Printf("running: %s (args: %v)", cmd.Path, cmd.Args[1:])

	if _, err := execCmd(cmd); err != nil {
		return errors.E(err, "failed to mount volume")
	}

	// the LTFS process goes to the background, not reporting the PID. So, do
	// some evil stuff and find it.
	out, err := exec.Command("pgrep", "-f", fmt.Sprintf("ltfs %s", f.mountpath)).Output()
	if err != nil {
		return errors.E(err, "pgrep failed")
	}

	pid, err := strconv.Atoi(string(out[:len(out)-1]))
	if err != nil {
		return errors.E(err, "strconv.Atoi failed")
	}

	// On Unix systems, FindProcess always succeeds and returns
	// an *os.Process for the given pid, regardless of whether the
	// process exists. We know it exists because we just found it.
	f.proc, _ = os.FindProcess(pid)

	// a mounted LTFS file system is just an fsdir.Storage.
	f.Storage = fsdir.New(f.mountpath)

	return nil
}

// Unmount unmounts the LTFS mounted at path and waits for the process
// identified by proc to terminate.
func (f *impl) Unmount() error {
	cmd := exec.Command(fusermountCommand, f.mountpath)

	_, err := execCmd(cmd)
	if err != nil {
		return errors.E(err, "failed to execute fusermount")
	}

	state, err := f.proc.Wait()
	if err != nil {
		return err
	}

	if !state.Success() {
		status := state.Sys().(syscall.WaitStatus)
		return errors.E(errors.Internal, errors.Strf("ltfs process exited with status %d", status.ExitStatus()))
	}

	f.Storage = nil

	return nil
}

func execCmd(cmd *exec.Cmd) ([]byte, error) {
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
