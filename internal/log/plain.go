package log

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/alecthomas/types/once"
	"github.com/mattn/go-isatty"
)

var (
	// https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
	// Colours for modules.
	rgb256Colours = []string{
		"000000", "008000", "808000", "000080", "800080", "008080", "c0c0c0", "808080",
		"00ff00", "ffff00", "0000ff", "ff00ff", "00ffff", "ffffff", "000000", "00005f", "000087", "0000af",
		"0000d7", "0000ff", "005f00", "005f5f", "005f87", "005faf", "005fd7", "005fff", "008700", "00875f",
		"008787", "0087af", "0087d7", "0087ff", "00af00", "00af5f", "00af87", "00afaf", "00afd7", "00afff",
		"00d700", "00d75f", "00d787", "00d7af", "00d7d7", "00d7ff", "00ff00", "00ff5f", "00ff87", "00ffaf",
		"00ffd7", "00ffff", "5f0000", "5f005f", "5f0087", "5f00af", "5f00d7", "5f00ff", "5f5f00", "5f5f5f",
		"5f5f87", "5f5faf", "5f5fd7", "5f5fff", "5f8700", "5f875f", "5f8787", "5f87af", "5f87d7", "5f87ff",
		"5faf00", "5faf5f", "5faf87", "5fafaf", "5fafd7", "5fafff", "5fd700", "5fd75f", "5fd787", "5fd7af",
		"5fd7d7", "5fd7ff", "5fff00", "5fff5f", "5fff87", "5fffaf", "5fffd7", "5fffff", "870000", "87005f",
		"870087", "8700af", "8700d7", "8700ff", "875f00", "875f5f", "875f87", "875faf", "875fd7", "875fff",
		"878700", "87875f", "878787", "8787af", "8787d7", "8787ff", "87af00", "87af5f", "87af87", "87afaf",
		"87afd7", "87afff", "87d700", "87d75f", "87d787", "87d7af", "87d7d7", "87d7ff", "87ff00", "87ff5f",
		"87ff87", "87ffaf", "87ffd7", "87ffff", "af0000", "af005f", "af0087", "af00af", "af00d7", "af00ff",
		"af5f00", "af5f5f", "af5f87", "af5faf", "af5fd7", "af5fff", "af8700", "af875f", "af8787", "af87af",
		"af87d7", "af87ff", "afaf00", "afaf5f", "afaf87", "afafaf", "afafd7", "afafff", "afd700", "afd75f",
		"afd787", "afd7af", "afd7d7", "afd7ff", "afff00", "afff5f", "afff87", "afffaf", "afffd7", "afffff",
		"d70000", "d7005f", "d70087", "d700af", "d700d7", "d700ff", "d75f00", "d75f5f", "d75f87", "d75faf",
		"d75fd7", "d75fff", "d78700", "d7875f", "d78787", "d787af", "d787d7", "d787ff", "d7af00", "d7af5f",
		"d7af87", "d7afaf", "d7afd7", "d7afff", "d7d700", "d7d75f", "d7d787", "d7d7af", "d7d7d7", "d7d7ff",
		"d7ff00", "d7ff5f", "d7ff87", "d7ffaf", "d7ffd7", "d7ffff", "ff0000", "ff005f", "ff0087", "ff00af",
		"ff00d7", "ff00ff", "ff5f00", "ff5f5f", "ff5f87", "ff5faf", "ff5fd7", "ff5fff", "ff8700", "ff875f",
		"ff8787", "ff87af", "ff87d7", "ff87ff", "ffaf00", "ffaf5f", "ffaf87", "ffafaf", "ffafd7", "ffafff",
		"ffd700", "ffd75f", "ffd787", "ffd7af", "ffd7d7", "ffd7ff", "ffff00", "ffff5f", "ffff87", "ffffaf",
		"ffffd7", "ffffff", "080808", "121212", "1c1c1c", "262626", "303030", "3a3a3a", "444444", "4e4e4e",
		"585858", "606060", "666666", "767676", "808080", "8a8a8a", "949494", "9e9e9e", "a8a8a8", "b2b2b2",
		"bcbcbc", "c6c6c6", "d0d0d0", "dadada", "e4e4e4", "eeeeee",
	}
	scopeColours = once.Once(func(ctx context.Context) ([]string, error) {
		out := make([]string, 0, len(rgb256Colours))
		for i, c := range rgb256Colours {
			r, g, b := parseRGB(c)
			if (r > g*2 && r > b*2) || (r+g+b < 200) || (r == g && g == b) {
				continue
			}
			out = append(out, fmt.Sprintf("\033[38;5;%dm", i))
		}
		return out, nil
	})
	levelColours = map[Level]string{
		Trace: "\x1b[90m", // Dark gray
		Debug: "\x1b[34m", // Blue
		Info:  "\x1b[37m", // White
		Warn:  "\x1b[33m", // Yellow
		Error: "\x1b[31m", // Red
	}
)

var _ Sink = (*plainSink)(nil)

func newPlainSink(w io.Writer, logTime bool, alwaysColor bool) *plainSink {
	var isaTTY bool
	if alwaysColor {
		isaTTY = true
	} else if f, ok := w.(*os.File); ok {
		isaTTY = isatty.IsTerminal(f.Fd())
	}
	return &plainSink{
		isaTTY:  isaTTY,
		w:       w,
		logTime: logTime,
	}
}

type plainSink struct {
	isaTTY  bool
	w       io.Writer
	logTime bool
}

// Log implements Sink
func (t *plainSink) Log(entry Entry) error {
	var prefix string

	// Add timestamp if required
	if t.logTime {
		prefix += entry.Time.Format(time.StampMilli) + " "
	}

	// Add scope if required
	scopeString := ""
	scope, exists := entry.Attributes[scopeKey]
	module, moduleExists := entry.Attributes[moduleKey]
	if moduleExists {
		if exists && scope != module {
			scopeString = module + ":" + scope
		} else {
			scopeString = module
		}
	} else if exists {
		scopeString = scope
		module = scope
	}
	if exists {
		if t.isaTTY {
			prefix += entry.Level.String() + ":" + ScopeColor(module) + scopeString + "\x1b[0m: "
		} else {
			prefix += entry.Level.String() + ":" + scopeString + ": "
		}
	} else {
		prefix += entry.Level.String() + ": "
	}

	// Print
	var err error
	if t.isaTTY {
		_, err = fmt.Fprintf(t.w, "%s%s%s%s\x1b[0m\n", levelColours[entry.Level], prefix, levelColours[entry.Level], entry.Message)
	} else {
		_, err = fmt.Fprintf(t.w, "%s%s\n", prefix, entry.Message)
	}
	if err != nil {
		return err
	}
	return nil
}

//nolint:errcheck
func parseRGB(s string) (int, int, int) {
	r, _ := strconv.ParseInt(s[:2], 16, 32)
	g, _ := strconv.ParseInt(s[2:4], 16, 32)
	b, _ := strconv.ParseInt(s[4:], 16, 32)
	return int(r), int(g), int(b)
}

func ScopeColor(scope string) string {
	hash := fnv.New32a()
	hash.Write([]byte(scope))
	colours, _ := scopeColours.Get(context.Background()) //nolint:errcheck
	return colours[hash.Sum32()%uint32(len(colours))]
}
