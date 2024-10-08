package schema

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/internal/duration"
	"github.com/TBD54566975/ftl/internal/slices"
)

const (
	DefaultRetryCount  = 10
	MinBackoffLimitStr = "1s"
	MinBackoffLimit    = 1 * time.Second
	DefaultMaxBackoff  = 1 * time.Hour
	MaxBackoffLimitStr = "1d"
	MaxBackoffLimit    = 24 * time.Hour
)

//protobuf:6
type MetadataRetry struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Count      *int   `parser:"'+' 'retry' (@Number Whitespace)?" protobuf:"2,optional"`
	MinBackoff string `parser:"@(Number (?! Whitespace) Ident)?" protobuf:"3"`
	MaxBackoff string `parser:"@(Number (?! Whitespace) Ident)?" protobuf:"4"`
	Catch      *Ref   `parser:"('catch' @@)?" protobuf:"5,optional"`
}

var _ Metadata = (*MetadataRetry)(nil)

func (*MetadataRetry) schemaMetadata()      {}
func (m *MetadataRetry) Position() Position { return m.Pos }

func (m *MetadataRetry) schemaChildren() []Node {
	if m.Catch == nil {
		return nil
	}
	return []Node{m.Catch}
}

func (m *MetadataRetry) String() string {
	components := []string{"+retry"}
	if m.Count != nil {
		components = append(components, strconv.Itoa(*m.Count))
	}
	components = append(components, m.MinBackoff)
	if len(m.MaxBackoff) > 0 {
		components = append(components, m.MaxBackoff)
	}
	if m.Catch != nil {
		components = append(components, fmt.Sprintf("catch %v", m.Catch))
	}
	return strings.Join(components, " ")
}

func (m *MetadataRetry) ToProto() proto.Message {
	var count *int64
	if m.Count != nil {
		count = proto.Int64(int64(*m.Count))
	}
	var catch *schemapb.Ref
	if m.Catch != nil {
		catch = m.Catch.ToProto().(*schemapb.Ref) //nolint: forcetypeassert
	}
	return &schemapb.MetadataRetry{
		Pos:        posToProto(m.Pos),
		Count:      count,
		MinBackoff: m.MinBackoff,
		MaxBackoff: m.MaxBackoff,
		Catch:      catch,
	}
}

func parseRetryDuration(str string) (time.Duration, error) {
	dur, err := duration.Parse(str)
	if err != nil {
		return 0, fmt.Errorf("could not parse retry duration: %w", err)
	}
	if dur < MinBackoffLimit {
		return 0, fmt.Errorf("retry must have a minimum backoff of %v", MinBackoffLimitStr)
	}
	if dur > MaxBackoffLimit {
		return 0, fmt.Errorf("retry backoff can not be larger than %v", MaxBackoffLimitStr)
	}
	return dur, nil
}

type RetryParams struct {
	Count      int
	MinBackoff time.Duration
	MaxBackoff time.Duration
	Catch      optional.Option[RefKey]
}

func (m *MetadataRetry) RetryParams() (RetryParams, error) {
	params := RetryParams{}
	// count
	if m.Count != nil {
		params.Count = *m.Count
	} else {
		params.Count = DefaultRetryCount
	}

	// min backoff
	if m.MinBackoff == "" {
		if params.Count != 0 {
			return RetryParams{}, fmt.Errorf("retry must have a minimum backoff")
		}
		params.MinBackoff = MinBackoffLimit
	} else {
		minBackoff, err := parseRetryDuration(m.MinBackoff)
		if err != nil {
			return RetryParams{}, fmt.Errorf("could not parse min backoff duration: %w", err)
		}
		params.MinBackoff = minBackoff
	}

	// max backoff
	if m.MaxBackoff == "" {
		params.MaxBackoff = max(params.MinBackoff, DefaultMaxBackoff)
	} else {
		maxBackoff, err := parseRetryDuration(m.MaxBackoff)
		if err != nil {
			return RetryParams{}, fmt.Errorf("could not parse max backoff duration: %w", err)
		}
		params.MaxBackoff = maxBackoff
	}
	if m.Catch != nil {
		params.Catch = optional.Some(m.Catch.ToRefKey())
	}
	return params, nil
}

// RetryParamsForFSMTransition finds the retry metadata for the given transition and returns the retry count, min and max backoff.
func RetryParamsForFSMTransition(fsm *FSM, verb *Verb) (RetryParams, error) {
	// Find retry metadata
	retryMetadata, ok := slices.FindVariant[*MetadataRetry](verb.Metadata)
	if !ok {
		// default to fsm's retry metadata
		retryMetadata, ok = slices.FindVariant[*MetadataRetry](fsm.Metadata)
		if !ok {
			// no retry
			return RetryParams{}, nil
		}
	}

	return retryMetadata.RetryParams()
}
