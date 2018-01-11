// +build linux

package ltfs

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"hpt.space/tapr/errors"
)

const (
	ltfsCommand       = "/usr/local/bin/ltfs"
	fusermountCommand = "/usr/bin/fusermount"
	mkltfsCommand     = "/usr/local/bin/mkltfs"
)

// Format formats the volume currently in the drive at the given
// device path.
func Format(devpath string) error {
	cmd := exec.Command(mkltfsCommand, fmt.Sprintf("-d %s", devpath))

	_, err := execCmd(cmd)
	if err != nil {
		return errors.E(err, "failed to format volume")
	}

	return nil
}

// Mount mounts the LTFS formatted volume loaded in the device pointed
// to by devpath into the path directory. The function returns
// an *os.Process refering to the background ltfs process responsible
// for the mounted file system.
func Mount(devpath, path string) (*os.Process, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return nil, errors.E(errors.Invalid, errors.Strf("%s does not exist", path))
	}

	opts := []string{
		path,

		fmt.Sprintf("-o devname=%s", devpath),
		fmt.Sprintf("-o sync_type=%s", "unmount"),
	}

	cmd := exec.Command(ltfsCommand, opts...)

	if _, err = execCmd(cmd); err != nil {
		return nil, errors.E(err, "failed to mount volume")
	}

	// the LTFS process goes to the background, not reporting the PID. So, do
	// some evil stuff and find it.
	out, err := exec.Command("pgrep", "-f", fmt.Sprintf("ltfs %s", path)).Output()
	if err != nil {
		return nil, errors.E(err, "pgrep failed")
	}

	pid, err := strconv.Atoi(string(out[:len(out)-1]))
	if err != nil {
		return nil, errors.E(err, "strconv.Atoi failed")
	}

	// On Unix systems, FindProcess always succeeds and returns
	// an *os.Process for the given pid, regardless of whether the
	// process exists. We know it exists because we just found it.
	proc, _ := os.FindProcess(pid)

	return proc, nil
}

// Unmount unmounts the LTFS mounted at path and waits for the process
// identified by proc to terminate.
func Unmount(path string, proc *os.Process) error {
	cmd := exec.Command(fusermountCommand, path)

	_, err := execCmd(cmd)
	if err != nil {
		return errors.E(err, "failed to execute fusermount")
	}

	state, err := proc.Wait()
	if err != nil {
		return err
	}

	if !state.Success() {
		status := state.Sys().(syscall.WaitStatus)
		return errors.E(errors.Internal, errors.Strf("ltfs process exited with status %d", status.ExitStatus()))
	}

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
