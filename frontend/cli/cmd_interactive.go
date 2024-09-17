package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/chzyer/readline"
	"github.com/kballard/go-shellquote"

	"github.com/TBD54566975/ftl/internal/projectconfig"
)

type interactiveCmd struct {
}

func (i *interactiveCmd) Run(ctx context.Context, k *kong.Kong, projectConfig projectconfig.Config) error {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[32m>\033[0m ",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return fmt.Errorf("init readline: %w", err)
	}
	l.CaptureExitSignal()
	for {
		line, err := l.Readline()
		if errors.Is(err, readline.ErrInterrupt) {
			if len(line) == 0 {
				break
			}
			continue
		} else if errors.Is(err, io.EOF) {
			break
		}
		line = strings.TrimSpace(line)
		args, err := shellquote.Split(line)
		if err != nil {
			errorf("%s", err)
			continue
		}
		kctx, err := k.Parse(args)
		if err != nil {
			errorf("%s", err)
			continue
		}
		subctx := bindContext(ctx, kctx, projectConfig)

		err = kctx.Run(subctx)
		if err != nil {
			errorf("%s", err)
			continue
		}
	}
	return nil
}

func errorf(format string, args ...any) {
	fmt.Printf("\033[31m%s\033[0m\n", fmt.Sprintf(format, args...))
}
