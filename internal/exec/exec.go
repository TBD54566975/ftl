package exec

import (
	"context"
	"os"
	"os/exec"
	"syscall"

	"github.com/alecthomas/errors"
	"github.com/kballard/go-shellquote"

	"github.com/TBD54566975/ftl/internal/log"
)

type Cmd struct {
	*exec.Cmd
}

func LookPath(exe string) (string, error) {
	path, err := exec.LookPath(exe)
	return path, errors.WithStack(err)
}

func Command(ctx context.Context, dir, exe string, args ...string) *Cmd {
	logger := log.FromContext(ctx)
	pgid, err := syscall.Getpgid(0)
	if err != nil {
		panic(err)
	}
	logger.Tracef("exec: cd %s && %s %s", shellquote.Join(dir), exe, shellquote.Join(args...))
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pgid:    pgid,
		Setpgid: true,
	}
	cmd.Dir = dir
	output := logger.WriterAt(log.Debug)
	cmd.Stdout = output
	cmd.Stderr = output
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
