package routers

import (
	"context"
	"net/url"

	"github.com/TBD54566975/ftl/internal/configuration"
)

var _ configuration.Router[configuration.Secrets] = (*NoopRouter[configuration.Secrets])(nil)

type NoopRouter[R configuration.Role] struct {
}

func (f *NoopRouter[R]) Get(ctx context.Context, ref configuration.Ref) (key *url.URL, err error) {
	return nil, nil
}

func (f *NoopRouter[R]) List(ctx context.Context) ([]configuration.Entry, error) {
	out := make([]configuration.Entry, 0)
	return out, nil
}

func (f *NoopRouter[R]) Role() (role R) { return }

func (f *NoopRouter[R]) Set(ctx context.Context, ref configuration.Ref, key *url.URL) error {
	return nil
}

func (f *NoopRouter[R]) Unset(ctx context.Context, ref configuration.Ref) error {
	return nil
}
