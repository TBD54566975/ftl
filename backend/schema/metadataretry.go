package schema

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

const (
	MinBackoffLimitStr = "1s"
	MinBackoffLimit    = 1 * time.Second
	MaxBackoffLimitStr = "1d"
	MaxBackoffLimit    = 24 * time.Hour
)

type MetadataRetry struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Count      *int   `parser:"'+' 'retry' (@Number Whitespace)?" protobuf:"2,optional"`
	MinBackoff string `parser:"@(Number Ident)?" protobuf:"3"`
	MaxBackoff string `parser:"@(Number Ident)?" protobuf:"4"`
}

var _ Metadata = (*MetadataRetry)(nil)

func (*MetadataRetry) schemaMetadata()          {}
func (m *MetadataRetry) schemaChildren() []Node { return nil }
func (m *MetadataRetry) Position() Position     { return m.Pos }
func (m *MetadataRetry) String() string {
	components := []string{"+retry"}
	if m.Count != nil {
		components = append(components, strconv.Itoa(*m.Count))
	}
	components = append(components, m.MinBackoff)
	if len(m.MaxBackoff) > 0 {
		components = append(components, m.MaxBackoff)
	}
	return strings.Join(components, " ")
}

func (m *MetadataRetry) ToProto() proto.Message {
	count := int64(*m.Count)
	return &schemapb.MetadataRetry{
		Pos:        posToProto(m.Pos),
		Count:      &count,
		MinBackoff: m.MinBackoff,
		MaxBackoff: m.MaxBackoff,
	}
}

func (m *MetadataRetry) MinBackoffDuration() (time.Duration, error) {
	if m.MinBackoff == "" {
		return 0, fmt.Errorf("retry must have a minimum backoff")
	}
	duration, err := parseRetryDuration(m.MinBackoff)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

func (m *MetadataRetry) MaxBackoffDuration() (optional.Option[time.Duration], error) {
	if m.MaxBackoff == "" {
		return optional.None[time.Duration](), nil
	}
	duration, err := parseRetryDuration(m.MaxBackoff)
	if err != nil {
		return optional.None[time.Duration](), err
	}
	return optional.Some(duration), nil
}

func parseRetryDuration(str string) (time.Duration, error) {
	// regex is more lenient than what is valid to allow for better error messages.
	re := regexp.MustCompile(`(\d+)([a-zA-Z]+)`)

	var duration time.Duration
	previousUnitDuration := time.Duration(0)
	for len(str) > 0 {
		matches := re.FindStringSubmatchIndex(str)
		if matches == nil {
			return 0, fmt.Errorf("unable to parse retry backoff %q. Expected duration in format like '1m' or '30s'", str)
		}
		num, err := strconv.Atoi(str[matches[2]:matches[3]])
		if err != nil {
			return 0, fmt.Errorf("unable to parse retry backoff text %q: %w", str, err)
		}

		unitStr := str[matches[4]:matches[5]]
		var unitDuration time.Duration
		switch unitStr {
		case "d":
			unitDuration = time.Hour * 24
		case "h":
			unitDuration = time.Hour
		case "m":
			unitDuration = time.Minute
		case "s":
			unitDuration = time.Second
		default:
			return 0, fmt.Errorf("retry has unknown unit %q. Use 'd', 'h', 'm' or 's'", unitStr)
		}
		if previousUnitDuration != 0 && previousUnitDuration <= unitDuration {
			return 0, fmt.Errorf("retry has unit %q out of order. Units need to be ordered from largest to smallest", unitStr)
		}
		previousUnitDuration = unitDuration
		duration += time.Duration(num) * unitDuration
		str = str[matches[1]:]
	}
	if duration < MinBackoffLimit {
		return 0, fmt.Errorf("retry must have a minimum backoff of %v", MinBackoffLimitStr)
	}
	if duration > MaxBackoffLimit {
		return 0, fmt.Errorf("retry backoff can not be larger than %v", MaxBackoffLimitStr)
	}
	return duration, nil
}
