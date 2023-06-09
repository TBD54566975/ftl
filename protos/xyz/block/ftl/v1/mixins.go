package ftlv1

import (
	"strings"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	"github.com/oklog/ulid/v2"
)

func (m *Metadata) Set(key, value string) {
	out := make([]*Metadata_Pair, 0, len(m.Values))
	for _, pair := range m.Values {
		if !strings.EqualFold(pair.Key, key) {
			out = append(out, &Metadata_Pair{Key: pair.Key, Value: pair.Value})
		}
	}
	out = append(out, &Metadata_Pair{Key: key, Value: value})
	m.Values = out
}

func (m *Metadata) Add(key, value string) {
	m.Values = append(m.Values, &Metadata_Pair{Key: key, Value: value})
}

func (m *Metadata) Get(key string) types.Option[string] {
	for _, pair := range m.Values {
		if strings.EqualFold(pair.Key, key) {
			return types.Some(pair.Value)
		}
	}
	return types.None[string]()
}

func (m *Metadata) GetAll(key string) (out []string) {
	for _, pair := range m.Values {
		if strings.EqualFold(pair.Key, key) {
			out = append(out, pair.Value)
		}
	}
	return
}

func (m *Metadata) Delete(key string) {
	out := make([]*Metadata_Pair, 0, len(m.Values))
	for _, pair := range m.Values {
		if !strings.EqualFold(pair.Key, key) {
			out = append(out, pair)
		}
	}
	m.Values = out
}

func (r *RegisterRunnerRequest) DeploymentAsOptional() (types.Option[ulid.ULID], error) {
	if r.Deployment == nil {
		return types.None[ulid.ULID](), nil
	}
	key, err := ulid.Parse(*r.Deployment)
	if err != nil {
		return types.None[ulid.ULID](), errors.Wrap(err, "invalid deployment key")
	}
	return types.Some(key), nil
}
