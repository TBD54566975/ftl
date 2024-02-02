package exec

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/kballard/go-shellquote"

	"github.com/TBD54566975/ftl/backend/common/log"
)

type Cmd struct {
	*exec.Cmd
	level log.Level
}

func LookPath(exe string) (string, error) {
	path, err := exec.LookPath(exe)
	return path, err
}

func Capture(ctx context.Context, dir, exe string, args ...string) ([]byte, error) {
	cmd := Command(ctx, log.Debug, dir, exe, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	out, err := cmd.CombinedOutput()
	return out, err
}

func Command(ctx context.Context, level log.Level, dir, exe string, args ...string) *Cmd {
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
	output := logger.WriterAt(level)
	cmd.Stdout = output
	cmd.Stderr = output
	cmd.Env = os.Environ()
	return &Cmd{cmd, level}
}

// RunBuffered runs the command and captures the output. If the command fails, the output is logged.
func (c *Cmd) RunBuffered(ctx context.Context) error {
	outputBuffer := NewCircularBuffer(100)
	output := outputBuffer.WriterAt(ctx, c.level)
	c.Cmd.Stdout = output
	c.Cmd.Stderr = output

	err := c.Run()
	if err != nil {
		fmt.Printf("%s", outputBuffer.Bytes())
		return err
	}

	return nil
}

// Kill sends a signal to the process group of the command.
func (c *Cmd) Kill(signal syscall.Signal) error {
	if c.Process == nil {
		return nil
	}
	return syscall.Kill(c.Process.Pid, signal)
}
