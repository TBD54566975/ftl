package status

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
	"golang.org/x/term"
)

type BuildState string

const BuildStateWaiting BuildState = "\u001B[93mwaiting"
const BuildStateBuilding BuildState = "\u001B[94mbuilding"
const BuildStateBuilt BuildState = "\u001B[92mbuilt"
const BuildStateDeploying BuildState = "\u001B[94mdeploying"
const BuildStateDeployed BuildState = "\u001B[92mdeployed"
const BuildStateFailed BuildState = "\u001B[91mfailed"
const buildStateMaxLength = len(BuildStateDeploying)

var _ StatusManager = &terminalStatusManager{}
var _ StatusLine = &terminalStatusLine{}

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
	sm := &terminalStatusManager{statusLock: sync.RWMutex{}, moduleStates: map[string]BuildState{}, height: height, width: width}
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
		for !sm.closed.Load() {
			buf := bytes.Buffer{}
			rawData := make([]byte, 104)
			n, err := sm.read.Read(rawData)
			if err != nil {
				// Not much we can do here
				sm.Close()
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

func (r *terminalStatusManager) gotoCoords(line int, col int) {
	r.underlyingWrite(fmt.Sprintf("\033[%d;%dH", line, col))
}

func (r *terminalStatusManager) gotoLine(line int) {
	r.gotoCoords(line, 0)
}
func (r *terminalStatusManager) clearStatusMessages() {
	if r.totalStatusLines == 0 {
		return
	}
	r.gotoLine(r.height - r.totalStatusLines + 1)
	r.underlyingWrite("\033[J")
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
	r.clearStatusMessages()
	r.underlyingWrite("\n")
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
	r.closed.Store(true)
	_ = r.write.Close()
	r.clearStatusMessages()
	os.Stdout = r.old // restoring the real stdout
	os.Stderr = r.oldErr
}
func (r *terminalStatusManager) writeLine(s string) {
	if r.height < 7 || r.width < 20 || r.height-r.totalStatusLines < 5 {
		// Not enough space to draw anything
		r.underlyingWrite(s + "\n")
		return
	}
	r.statusLock.RLock()
	defer r.statusLock.RUnlock()

	if r.totalStatusLines == 0 {
		r.underlyingWrite("\n" + s)
		return
	}
	r.clearStatusMessages()
	r.gotoLine(r.height - r.totalStatusLines)
	r.underlyingWrite("\n" + s)
	for range r.totalStatusLines {
		r.underlyingWrite("\n")
	}
	if r.totalStatusLines == 0 {
		r.underlyingWrite("\n")
	}
	r.gotoLine(r.height)
	r.redrawStatus()

}
func (r *terminalStatusManager) redrawStatus() {
	if r.height < 7 || r.width < 20 || r.height-r.totalStatusLines < 5 {
		// Not enough space to draw anything
		return
	}

	if r.totalStatusLines == 0 || r.closed.Load() {
		return
	}
	// If the console is tiny we don't do this
	r.clearStatusMessages()
	r.gotoLine(r.height - r.totalStatusLines)

	r.underlyingWrite("\n--\n")
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

}

func (r *terminalStatusManager) recalculateLines() {

	if len(r.moduleStates) > 0 && r.moduleLine != nil {

		entryLength := 0
		keys := []string{}
		for k := range r.moduleStates {
			// We use the max length rather than the actual length to avoid flickering
			// ANSI control characters are 5 bytes long
			thisLength := len(k) + buildStateMaxLength + 4 - 5 // 4 is the length of the ": " and the two trailing spaces
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
		for i, k := range keys {
			if i%perLine == 0 {
				msg += "\n"
			}
			pad := strings.Repeat(" ", entryLength-len(k)-len(r.moduleStates[k])-4+5)
			msg += "\u001B[97m" + k + ": " + string(r.moduleStates[k]) + "\u001B[39m  " + pad
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

func (r *noopStatusManager) Close() {

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
