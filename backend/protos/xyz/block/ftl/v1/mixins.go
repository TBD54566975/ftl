package ftlv1

import (
	"strings"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/controller/state"
)

func ArtefactToProto(artefact *state.DeploymentArtefact) *DeploymentArtefact {
	return &DeploymentArtefact{
		Path:       artefact.Path,
		Executable: artefact.Executable,
		Digest:     artefact.Digest.String(),
	}
}

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

func (m *Metadata) Get(key string) optional.Option[string] {
	for _, pair := range m.Values {
		if strings.EqualFold(pair.Key, key) {
			return optional.Some(pair.Value)
		}
	}
	return optional.None[string]()
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
