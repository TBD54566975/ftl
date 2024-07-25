package flock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/sys/unix"
)

var ErrLocked = errors.New("locked")

// Acquire a lock on the given path.
//
// The lock is released when the returned function is called.
func Acquire(ctx context.Context, path string, timeout time.Duration) (release func() error, err error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	end := time.Now().Add(timeout)
	for {
		release, err := acquire(absPath)
		if err == nil {
			return release, nil
		}
		if !errors.Is(err, ErrLocked) {
			return nil, fmt.Errorf("failed to acquire lock %s: %w", absPath, err)
		}
		if time.Now().After(end) {
			pid, _ := os.ReadFile(absPath) //nolint:errcheck
			return nil, fmt.Errorf("timed out acquiring lock %s, locked by pid %s: %w", absPath, pid, err)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func acquire(path string) (release func() error, err error) {
	pid := os.Getpid()
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_SYNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("open failed: %w", err)
	}

	err = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)
	if err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("%w: %w", ErrLocked, err)
	}

	_, err = unix.Write(fd, []byte(strconv.Itoa(pid)))
	if err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}
	return func() error {
		return errors.Join(unix.Flock(fd, unix.LOCK_UN), unix.Close(fd), os.Remove(path))
	}, nil
}
