package async

import (
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

type asyncOriginParseRoot struct {
	Key AsyncOrigin `parser:"@@"`
}

var asyncOriginLexer = lexer.MustSimple([]lexer.SimpleRule{
	{"NumberIdent", `[0-9][a-zA-Z0-9_-]*`},
	{"Ident", `[a-zA-Z_][a-zA-Z0-9_-]*`},
	{"Punct", `[:.]`},
})

var asyncOriginParser = participle.MustBuild[asyncOriginParseRoot](
	participle.Union[AsyncOrigin](AsyncOriginCron{}, AsyncOriginPubSub{}),
	participle.Lexer(asyncOriginLexer),
)

// AsyncOrigin is a sum type representing the originator of an async call.
//
// This is used to determine how to handle the result of the async call.
type AsyncOrigin interface {
	asyncOrigin()
	// Origin returns the origin type.
	Origin() string
	String() string
}

// AsyncOriginCron represents the context for the originator of a cron async call.
//
// It is in the form cron:<module>.<verb>
type AsyncOriginCron struct {
	CronJobKey model.CronJobKey `parser:"'cron' ':' @(~EOF)+"`
}

var _ AsyncOrigin = AsyncOriginCron{}

func (AsyncOriginCron) asyncOrigin()     {}
func (a AsyncOriginCron) Origin() string { return "cron" }
func (a AsyncOriginCron) String() string { return fmt.Sprintf("cron:%s", a.CronJobKey) }

// AsyncOriginPubSub represents the context for the originator of an PubSub async call.
//
// It is in the form sub:<module>.<subscription_name>
type AsyncOriginPubSub struct {
	Subscription schema.RefKey `parser:"'sub' ':' @@"`
}

var _ AsyncOrigin = AsyncOriginPubSub{}

func (AsyncOriginPubSub) asyncOrigin()     {}
func (a AsyncOriginPubSub) Origin() string { return "sub" }
func (a AsyncOriginPubSub) String() string { return fmt.Sprintf("sub:%s", a.Subscription) }

// ParseAsyncOrigin parses an async origin key.
func ParseAsyncOrigin(origin string) (AsyncOrigin, error) {
	root, err := asyncOriginParser.ParseString("", origin)
	if err != nil {
		return nil, fmt.Errorf("failed to parse async origin: %w", err)
	}
	return root.Key, nil
}
