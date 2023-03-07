package exec

import (
	"context"
	"os"
	"os/exec"
	"syscall"

	"github.com/alecthomas/errors"
)

type Cmd struct {
	*exec.Cmd
}

func LookPath(exe string) (string, error) {
	path, err := exec.LookPath(exe)
	return path, errors.WithStack(err)
}

func Command(ctx context.Context, dir, exe string, args ...string) *Cmd {
	pgid, err := syscall.Getpgid(0)
	if err != nil {
		panic(err)
	}
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pgid:    pgid,
		Setpgid: true,
	}
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return &Cmd{cmd}
}

// Kill sends a signal to the process group of the command.
func (c *Cmd) Kill(signal syscall.Signal) error {
	if c.Process == nil {
		return nil
	}
	return errors.WithStack(syscall.Kill(c.Process.Pid, signal))
}
