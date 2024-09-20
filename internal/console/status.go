package console

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"
	"unicode/utf8"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/kong"
	"github.com/tidwall/pretty"
	"golang.org/x/term"

	"github.com/TBD54566975/ftl/internal/projectconfig"
)

type BuildState string

const BuildStateWaiting BuildState = " 🚦️"
const BuildStateBuilding BuildState = " 🏗️"
const BuildStateBuilt BuildState = "📦️️"
const BuildStateDeploying BuildState = " 🚚️"
const BuildStateDeployed BuildState = " ✅️️"
const BuildStateFailed BuildState = "💥"

// moduleStatusPadding is the padding between module status entries
// it accounts for the colon, space and the emoji
const moduleStatusPadding = 10

var _ StatusManager = &terminalStatusManager{}
var _ StatusLine = &terminalStatusLine{}

var buildColors map[BuildState]string

func init() {
	buildColors = map[BuildState]string{
		BuildStateWaiting:   "\u001B[93m",
		BuildStateBuilding:  "\u001B[94m",
		BuildStateBuilt:     "\u001B[92m",
		BuildStateDeploying: "\u001B[94m",
		BuildStateDeployed:  "\u001B[92m",
		BuildStateFailed:    "\u001B[91m",
	}
}

type StatusManager interface {
	Close()
	NewStatus(message string) StatusLine
	IntoContext(ctx context.Context) context.Context
	SetModuleState(module string, state BuildState)
}

type StatusLine interface {
	SetMessage(message string)
	Close()
}

type terminalStatusManager struct {
	old              *os.File
	oldErr           *os.File
	read             *os.File
	write            *os.File
	closed           atomic.Value[bool]
	totalStatusLines int
	statusLock       sync.RWMutex
	lines            []*terminalStatusLine
	moduleLine       *terminalStatusLine
	moduleStates     map[string]BuildState
	height           int
	width            int
	exitWait         sync.WaitGroup
	console          bool
	consoleRefresh   func()
}

type statusKey struct{}

var statusKeyInstance = statusKey{}

func NewStatusManager(ctx context.Context) StatusManager {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return &noopStatusManager{}
	}
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return &noopStatusManager{}
	}
	sm := &terminalStatusManager{statusLock: sync.RWMutex{}, moduleStates: map[string]BuildState{}, height: height, width: width, exitWait: sync.WaitGroup{}}
	sm.exitWait.Add(1)
	sm.old = os.Stdout
	sm.oldErr = os.Stderr
	sm.read, sm.write, err = os.Pipe()

	if err != nil {
		return &noopStatusManager{}
	}
	os.Stdout = sm.write
	os.Stderr = sm.write

	go func() {
		for {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGWINCH)
			// Block until a signal is received.
			<-c
			sm.statusLock.Lock()
			sm.clearStatusMessages()

			sm.width, sm.height, _ = term.GetSize(int(sm.old.Fd())) //nolint:errcheck
			sm.recalculateLines()
			sm.statusLock.Unlock()
		}
	}()

	go func() {
		defer sm.exitWait.Done()
		current := ""
		for {

			buf := bytes.Buffer{}
			rawData := make([]byte, 104)
			n, err := sm.read.Read(rawData)
			if err != nil {
				if current != "" {
					sm.writeLine(current)
				}
				return
			}
			buf.Write(rawData[:n])
			for buf.Len() > 0 {
				d, s, err := buf.ReadRune()
				if err != nil {
					// EOF, need to read more data
					break
				}
				if d == utf8.RuneError && s == 1 {
					if buf.Available() < 4 && !sm.closed.Load() {
						_ = buf.UnreadByte() //nolint:errcheck
						// Need to read more data, probably not a full rune
						break
					}
					// Otherwise, ignore the error, not much we can do
					continue
				}
				if d == '\n' {
					sm.writeLine(current)
					current = ""
				} else {
					current += string(d)
				}
			}
		}
	}()

	go func() {
		<-ctx.Done()
		sm.Close()
	}()
	return sm
}

func UpdateModuleState(ctx context.Context, module string, state BuildState) {
	sm := FromContext(ctx)
	sm.SetModuleState(module, state)
}

// PrintJSON prints a json string to the terminal
// It probably doesn't belong here, but it will be moved later with the interactive terminal work
func PrintJSON(ctx context.Context, json []byte) {
	sm := FromContext(ctx)
	if _, ok := sm.(*terminalStatusManager); ok {
		// ANSI enabled
		fmt.Printf("%s\n", pretty.Color(pretty.Pretty(json), nil))
	} else {
		fmt.Printf("%s\n", json)
	}
}

func (r *terminalStatusManager) gotoCoords(line int, col int) {
	r.underlyingWrite(fmt.Sprintf("\033[%d;%dH", line, col))
}

func (r *terminalStatusManager) clearStatusMessages() {
	if r.totalStatusLines == 0 {
		return
	}
	count := r.totalStatusLines
	if r.console {
		count--
		// Don't clear the console line
		r.underlyingWrite("\u001B[1A")
	}
	for range count {
		r.underlyingWrite("\033[2K\u001B[1A")
	}
}

func (r *terminalStatusManager) IntoContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, statusKeyInstance, r)
}

func FromContext(ctx context.Context) StatusManager {
	sm, ok := ctx.Value(statusKeyInstance).(StatusManager)
	if !ok {
		return &noopStatusManager{}
	}
	return sm
}

func (r *terminalStatusManager) NewStatus(message string) StatusLine {
	r.statusLock.Lock()
	defer r.statusLock.Unlock()
	line := &terminalStatusLine{manager: r, message: message}
	r.newStatusInternal(line)
	return line
}

func (r *terminalStatusManager) newStatusInternal(line *terminalStatusLine) {
	if r.closed.Load() {
		return
	}
	for i, l := range r.lines {
		if l.priority < line.priority {
			r.lines = slices.Insert(r.lines, i, line)
			r.recalculateLines()
			return
		}
	}
	// If we get here we just append the line
	r.lines = append(r.lines, line)
	r.recalculateLines()
}

func (r *terminalStatusManager) SetModuleState(module string, state BuildState) {
	if module == "builtin" {
		return
	}
	r.statusLock.Lock()
	defer r.statusLock.Unlock()
	r.moduleStates[module] = state
	if r.moduleLine != nil {
		r.recalculateLines()
	} else {
		r.moduleLine = &terminalStatusLine{manager: r, priority: -10000}
		r.newStatusInternal(r.moduleLine)
	}
}

func (r *terminalStatusManager) Close() {
	r.statusLock.Lock()
	r.clearStatusMessages()
	r.totalStatusLines = 0
	r.lines = []*terminalStatusLine{}
	r.statusLock.Unlock()
	os.Stdout = r.old // restoring the real stdout
	os.Stderr = r.oldErr
	r.closed.Store(true)
	_ = r.write.Close() //nolint:errcheck
	r.exitWait.Wait()
}

func (r *terminalStatusManager) writeLine(s string) {
	r.statusLock.RLock()
	defer r.statusLock.RUnlock()

	if r.totalStatusLines == 0 {
		r.underlyingWrite("\n" + s)
		return
	}
	r.clearStatusMessages()
	r.underlyingWrite("\n" + s)
	r.redrawStatus()

}
func (r *terminalStatusManager) redrawStatus() {
	if r.totalStatusLines == 0 || r.closed.Load() {
		return
	}
	r.underlyingWrite("\n\n")
	for i := len(r.lines) - 1; i >= 0; i-- {
		msg := r.lines[i].message
		if msg != "" {
			r.underlyingWrite(msg)
			if i > 0 {
				// If there is any more messages to print we add a newline
				for j := range i {
					if r.lines[j].message != "" {
						r.underlyingWrite("\n")
						break
					}
				}
			}
		}
	}
	if r.console {
		r.underlyingWrite("\n")
	}
	if r.consoleRefresh != nil {
		r.consoleRefresh()
	}
}

func (r *terminalStatusManager) recalculateLines() {

	if len(r.moduleStates) > 0 && r.moduleLine != nil {
		entryLength := 0
		keys := []string{}
		for k := range r.moduleStates {
			thisLength := len(k) + moduleStatusPadding
			if thisLength > entryLength {
				entryLength = thisLength
			}
			keys = append(keys, k)
		}
		msg := ""
		perLine := r.width / entryLength
		if perLine == 0 {
			perLine = 1
		}
		slices.Sort(keys)
		multiLine := false
		for i, k := range keys {
			if i%perLine == 0 && i > 0 {
				msg += "\n"
				multiLine = true
			}
			pad := strings.Repeat(" ", entryLength-len(k)-moduleStatusPadding+2)
			state := r.moduleStates[k]
			msg += pad + buildColors[state] + k + ": " + string(state) + "\u001B[39m"
		}
		if !multiLine {
			// For multi-line messages we don't want to trim the message as we want to line up the columns
			// For a single line this just looks weird
			msg = strings.TrimSpace(msg)
		}
		r.moduleLine.message = msg
	}
	total := 0
	for _, i := range r.lines {
		if i.message != "" {
			total++
			total += countLines(i.message, r.width)
		}
	}
	if total > 0 {
		total++
	}
	if r.console {
		total++
	}
	r.clearStatusMessages()
	r.totalStatusLines = total
	r.redrawStatus()
}

func (r *terminalStatusManager) underlyingWrite(messages string) {
	_, _ = r.old.WriteString(messages) //nolint:errcheck
}

func countLines(s string, width int) int {
	return countLinesAtPos(s, 0, width)
}

func countLinesAtPos(s string, cursorPos int, width int) int {
	if s == "" {
		return 0
	}
	lines := 0
	curLength := cursorPos
	for i := range s {
		if s[i] == '\n' {
			lines++
			curLength = 0
		} else {
			curLength++
			if curLength == width {
				lines++
				curLength = 0
			}
		}
	}
	return lines
}

type terminalStatusLine struct {
	manager  *terminalStatusManager
	message  string
	priority int
}

func (r *terminalStatusLine) Close() {
	r.manager.statusLock.Lock()
	defer r.manager.statusLock.Unlock()

	r.manager.clearStatusMessages()
	for i := range r.manager.lines {
		if r.manager.lines[i] == r {
			r.manager.lines = append(r.manager.lines[:i], r.manager.lines[i+1:]...)
			r.manager.recalculateLines()
			return
		}
	}
	r.manager.redrawStatus()
}

func (r *terminalStatusLine) SetMessage(message string) {
	r.manager.statusLock.Lock()
	defer r.manager.statusLock.Unlock()
	r.message = message
	r.manager.recalculateLines()
}

func LaunchEmbeddedConsole(ctx context.Context, k *kong.Kong, projectConfig projectconfig.Config, binder KongContextBinder) {
	sm := FromContext(ctx)
	if tsm, ok := sm.(*terminalStatusManager); ok {
		tsm.console = true
		go func() {

			err := RunInteractiveConsole(ctx, k, projectConfig, binder, func(f func()) {
				tsm.statusLock.Lock()
				defer tsm.statusLock.Unlock()
				tsm.consoleRefresh = f
			})
			if err != nil {
				fmt.Printf("\033[31mError: %s\033[0m\n", err)
				return
			}
		}()
		tsm.statusLock.Lock()
		defer tsm.statusLock.Unlock()
		tsm.recalculateLines()
	}
}
