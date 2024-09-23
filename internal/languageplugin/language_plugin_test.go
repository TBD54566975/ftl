package languageplugin

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/assert/v2"
)

func TestLaunch(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	bind, err := url.Parse("http://127.0.0.1:36000")
	assert.NoError(t, err)
	p, err := New(ctx, "/Users/mtoohey/Code/ftl/examples/go/time", bind)
	assert.NoError(t, err)
	time.Sleep(20 * time.Second)
	fmt.Printf("plugin: %v\n", p)
	p.Kill()
}
