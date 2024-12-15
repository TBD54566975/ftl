package terminal

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
	"time"
	"unicode/utf8"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"github.com/tidwall/pretty"
	"golang.org/x/term"

	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

const ansiUpOneLine = "\u001B[1A"
const ansiClearLine = "\u001B[2K"
const ansiResetTextColor = "\u001B[39m"

type BuildState string

const BuildStateWaiting BuildState = "Waiting"
const BuildStateBuilding BuildState = "Building"
const BuildStateBuilt BuildState = "Built"
const BuildStateDeploying BuildState = "Deploying"
const BuildStateDeployed BuildState = "Deployed"
const BuildStateFailed BuildState = "Failed"
const BuildStateTerminated BuildState = "Terminated"

// moduleStatusPadding is the padding between module status entries
// it accounts for the icon, the module name, and the padding between them
const moduleStatusPadding = 5

var _ StatusManager = &terminalStatusManager{}
var _ StatusLine = &terminalStatusLine{}

var buildColors map[BuildState]string
var buildStateIcon map[BuildState]func(int) string

var spinner = []string{"◜", "◝", "◞", "◟"}

func init() {
	buildColors = map[BuildState]string{
		BuildStateWaiting:   "\u001B[93m",
		BuildStateBuilding:  "\u001B[94m",
		BuildStateBuilt:     "\u001B[92m",
		BuildStateDeploying: "\u001B[94m",
		BuildStateDeployed:  "\u001B[92m",
		BuildStateFailed:    "\u001B[91m",
	}
	spin := func(spinnerCount int) string {
		return spinner[spinnerCount]
	}
	block := func(int) string {
		return "✔"
	}
	cross := func(int) string {
		return "✘"
	}
	empty := func(int) string {
		return "•"
	}
	buildStateIcon = map[BuildState]func(int) string{
		BuildStateWaiting:   empty,
		BuildStateBuilding:  spin,
		BuildStateBuilt:     block,
		BuildStateDeploying: spin,
		BuildStateDeployed:  block,
		BuildStateFailed:    cross,
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
	old                *os.File
	oldErr             *os.File
	read               *os.File
	write              *os.File
	closed             atomic.Value[bool]
	totalStatusLines   int
	statusLock         sync.Mutex
	lines              []*terminalStatusLine
	moduleLine         *terminalStatusLine
	moduleStates       map[string]BuildState
	height             int
	width              int
	exitWait           sync.WaitGroup
	console            bool
	consoleRefresh     func()
	spinnerCount       int
	interactiveConsole optional.Option[*interactiveConsole]
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
	sm := &terminalStatusManager{statusLock: sync.Mutex{}, moduleStates: map[string]BuildState{}, height: height, width: width, exitWait: sync.WaitGroup{}}
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
		current := ""
		closed := false
		for {

			buf := bytes.Buffer{}
			rawData := make([]byte, 104)
			n, err := sm.read.Read(rawData)
			if err != nil {
				if current != "" {
					sm.writeLine(current, true)
				}
				if !closed {
					sm.clearStatusMessages()
					sm.exitWait.Done()
				}
				return
			}
			if closed {
				// When we are closed we just write the data to the old stdout
				_, _ = sm.old.Write(rawData[:n]) //nolint:errcheck
				continue
			}
			buf.Write(rawData[:n])
			for buf.Len() > 0 {
				d, s, err := buf.ReadRune()
				if d == 0 {
					// Null byte, we are done
					// we keep running though as there may be more data on exit
					// that we handle on a best effort basis
					sm.writeLine(current, true)
					if !closed {
						sm.clearStatusMessages()
						sm.exitWait.Done()
						closed = true
					}
					continue
				}
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
					sm.writeLine(current, false)
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

	// Animate the spinners

	go func() {
		for !sm.closed.Load() {
			time.Sleep(150 * time.Millisecond)
			sm.statusLock.Lock()
			if sm.spinnerCount == len(spinner)-1 {
				sm.spinnerCount = 0
			} else {
				sm.spinnerCount++
			}
			// only redraw if not stable
			stable := true
			for _, state := range sm.moduleStates {
				if state != BuildStateDeployed && state != BuildStateBuilt {
					stable = false
					break
				}
			}
			if !stable {
				sm.recalculateLines()
			}
			sm.statusLock.Unlock()

		}
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

func IsANSITerminal(ctx context.Context) bool {
	sm := FromContext(ctx)
	_, ok := sm.(*terminalStatusManager)
	return ok
}

func (r *terminalStatusManager) clearStatusMessages() {
	if r.totalStatusLines == 0 {
		return
	}
	count := r.totalStatusLines
	if r.console {
		count--
	}
	r.underlyingWrite(ansiClearLine)
	for range count {
		r.underlyingWrite(ansiUpOneLine + ansiClearLine)
	}
}

func (r *terminalStatusManager) consoleNewline(line string) {
	r.statusLock.Lock()
	defer r.statusLock.Unlock()
	count := r.totalStatusLines
	for range count {
		r.underlyingWrite(ansiUpOneLine + ansiClearLine)
	}
	r.underlyingWrite("\r" + line + "\n")
	if line == "" {
		r.redrawStatus()
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
	if state == BuildStateTerminated {
		delete(r.moduleStates, module)
	} else {
		r.moduleStates[module] = state
	}
	if r.moduleLine != nil {
		r.recalculateLines()
	} else {
		r.moduleLine = &terminalStatusLine{manager: r, priority: -10000}
		r.newStatusInternal(r.moduleLine)
	}
}

func (r *terminalStatusManager) Close() {
	r.statusLock.Lock()
	if it, ok := r.interactiveConsole.Get(); ok {
		it.Close()
	}
	r.clearStatusMessages()
	r.totalStatusLines = 0
	r.lines = []*terminalStatusLine{}
	r.statusLock.Unlock()
	os.Stdout = r.old // restoring the real stdout
	os.Stderr = r.oldErr
	r.closed.Store(true)
	// We send a null byte to the write pipe to unblock the read
	_, _ = r.write.Write([]byte{0}) //nolint:errcheck
	r.exitWait.Wait()
}

func (r *terminalStatusManager) writeLine(s string, last bool) {
	r.statusLock.Lock()
	defer r.statusLock.Unlock()
	if !last {
		s += "\n"
	}

	if r.totalStatusLines == 0 {
		r.underlyingWrite("\r" + s)
		return
	}
	r.clearStatusMessages()
	r.underlyingWrite("\r" + s)
	if !last {
		r.redrawStatus()
	}

}
func (r *terminalStatusManager) redrawStatus() {
	if r.statusLock.TryLock() {
		panic("redrawStatus called without holding the lock")
	}
	if r.totalStatusLines == 0 || r.closed.Load() {
		return
	}
	for i := len(r.lines) - 1; i >= 0; i-- {
		msg := r.lines[i].message
		if msg != "" {
			r.underlyingWrite("\r" + msg + "\n")
		}
	}
	if r.consoleRefresh != nil {
		r.consoleRefresh()
	}
}

func (r *terminalStatusManager) recalculateLines() {
	r.clearStatusMessages()
	total := 0
	if len(r.moduleStates) > 0 && r.moduleLine != nil {
		total++
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
				total++
			}
			pad := strings.Repeat(" ", entryLength-len(k)-moduleStatusPadding)
			state := r.moduleStates[k]
			msg += buildColors[state] + buildStateIcon[state](r.spinnerCount) + "[" + log.ScopeColor(k) + k + buildColors[state] + "]  " + ansiResetTextColor + pad
		}
		if !multiLine {
			// For multi-line messages we don't want to trim the message as we want to line up the columns
			// For a single line this just looks weird
			msg = strings.TrimSpace(msg)
		}
		r.moduleLine.message = msg
	}
	for _, i := range r.lines {
		if i.message != "" && i != r.moduleLine {
			total++
			total += countLines(i.message, r.width)
		}
	}
	if r.console {
		total++
	}
	r.totalStatusLines = total
	r.redrawStatus()
}

func (r *terminalStatusManager) underlyingWrite(messages string) {
	_, _ = r.old.WriteString(messages) //nolint:errcheck
}

func countLines(s string, width int) int {
	if s == "" {
		return 0
	}
	lines := 0
	curLength := 0
	// TODO: count unicode characters properly
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

func LaunchEmbeddedConsole(ctx context.Context, k *kong.Kong, binder KongContextBinder, eventSource schemaeventsource.EventSource) {
	sm := FromContext(ctx)
	if tsm, ok := sm.(*terminalStatusManager); ok {
		it, err := newInteractiveConsole(k, binder, eventSource)
		if err != nil {
			fmt.Printf("\033[31mError: %s\033[0m\n", err)
			return
		}
		tsm.interactiveConsole = optional.Some(it)
		go func() {
			err := it.run(ctx)
			if err != nil {
				fmt.Printf("\033[31mError: %s\033[0m\n", err)
				return
			}
		}()
	}
}
