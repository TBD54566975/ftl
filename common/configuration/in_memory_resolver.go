package configuration

import (
	"context"
	"fmt"
	"net/url"
)

type InMemoryResolver[R Role] struct {
	keyMap map[Ref]*url.URL
}

var _ Resolver[Configuration] = InMemoryResolver[Configuration]{}
var _ Resolver[Secrets] = InMemoryResolver[Secrets]{}

func NewInMemoryResolver[R Role]() *InMemoryResolver[R] {
	return &InMemoryResolver[R]{keyMap: map[Ref]*url.URL{}}
}

func (k InMemoryResolver[R]) Role() R { var r R; return r }

func (k InMemoryResolver[R]) Get(ctx context.Context, ref Ref) (*url.URL, error) {
	if key, found := k.keyMap[ref]; found {
		return key, nil
	}
	return nil, fmt.Errorf("key %q not found", ref.Name)
}

func (k InMemoryResolver[R]) List(ctx context.Context) ([]Entry, error) {
	entries := []Entry{}
	for ref, url := range k.keyMap {
		entries = append(entries, Entry{Ref: ref, Accessor: url})
	}
	return entries, nil
}

func (k InMemoryResolver[R]) Set(ctx context.Context, ref Ref, key *url.URL) error {
	k.keyMap[ref] = key
	return nil
}

func (k InMemoryResolver[R]) Unset(ctx context.Context, ref Ref) error {
	delete(k.keyMap, ref)
	return nil
}
