package console

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/chzyer/readline"
	kongcompletion "github.com/jotaen/kong-completion"
	"github.com/kballard/go-shellquote"
	"github.com/posener/complete"

	"github.com/TBD54566975/ftl/internal/projectconfig"
)

var _ readline.AutoCompleter = &FTLCompletion{}
var errExitTrap = errors.New("exit trap")

type KongContextBinder func(ctx context.Context, kctx *kong.Context, projectConfig projectconfig.Config, app *kong.Kong, cancel context.CancelFunc) context.Context

func RunInteractiveConsole(ctx context.Context, k *kong.Kong, projectConfig projectconfig.Config, binder KongContextBinder, refreshFunction func(func()), cancelContext context.CancelFunc) error {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[32m>\033[0m ",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete:    &FTLCompletion{app: k},
		Listener:        &ExitListener{cancel: cancelContext},
	})
	if refreshFunction != nil {
		refreshFunction(l.Refresh)
	}
	if err != nil {
		return fmt.Errorf("init readline: %w", err)
	}
	context.AfterFunc(ctx, func() {
		_ = l.Close()
	})
	l.CaptureExitSignal()
	// Overload the exit function to avoid exiting the process
	k.Exit = func(i int) { panic(errExitTrap) }
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
		if line == "" {
			continue
		}
		args, err := shellquote.Split(line)
		if err != nil {
			errorf("%s", err)
			continue
		}
		func() {
			defer func() {
				// Catch Exit() and continue the loop
				if r := recover(); r != nil {
					if r == errExitTrap { //nolint:errorlint
						return
					}
				}
			}()
			kctx, err := k.Parse(args)
			if err != nil {
				errorf("%s", err)
				return
			}
			subctx := binder(ctx, kctx, projectConfig, k, cancelContext)

			err = kctx.Run(subctx)
			if err != nil {
				errorf("error: %s", err)
				return
			}
			// Force a status refresh
			println("")
		}()
	}
	return nil
}

var _ readline.Listener = &ExitListener{}

type ExitListener struct {
	cancel context.CancelFunc
}

func (e ExitListener) OnChange(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
	if key == readline.CharInterrupt {
		e.cancel()
	}
	return line, pos, true
}

func errorf(format string, args ...any) {
	fmt.Printf("\033[31m%s\033[0m\n", fmt.Sprintf(format, args...))
}

type FTLCompletion struct {
	app *kong.Kong
}

func (f *FTLCompletion) Do(line []rune, pos int) ([][]rune, int) {
	parser := f.app
	if parser == nil {
		return nil, 0
	}
	all := []string{}
	completed := []string{}
	last := ""
	lastCompleted := ""
	lastSpace := false
	// We don't care about anything past pos
	// this completer can't handle completing in the middle of things
	if pos < len(line) {
		line = line[:pos]
	}
	current := 0
	for i, arg := range line {
		if i == pos {
			break
		}
		if arg == ' ' {
			lastWord := string(line[current:i])
			all = append(all, lastWord)
			completed = append(completed, lastWord)
			current = i + 1
			lastSpace = true
		} else {
			lastSpace = false
		}
	}
	if pos > 0 {
		if lastSpace {
			lastCompleted = all[len(all)-1]
		} else {
			if current < len(line) {
				last = string(line[current:])
				all = append(all, last)
			}
			if len(all) > 0 {
				lastCompleted = all[len(all)-1]
			}
		}
	}

	args := complete.Args{
		Completed:     completed,
		All:           all,
		Last:          last,
		LastCompleted: lastCompleted,
	}

	command, err := kongcompletion.Command(parser)
	if err != nil {
		// TODO handle error
		println(err.Error())
	}
	result := command.Predict(args)
	runes := [][]rune{}
	for _, s := range result {
		if !strings.HasPrefix(s, last) || s == "interactive" {
			continue
		}
		s = s[len(last):]
		str := []rune(s)
		runes = append(runes, str)
	}
	return runes, pos
}
