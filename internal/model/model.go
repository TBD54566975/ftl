package model

import (
	"strconv"
	"strings"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/sha256"
)

type Deployment struct {
	Module   string
	Language string
	Key      DeploymentKey
	Schema   *schema.Module
}

type Artefact struct {
	Path       string
	Executable bool
	Digest     sha256.SHA256
}

type Labels map[string]any

func (l Labels) String() string {
	w := strings.Builder{}
	i := 0
	for k, v := range l {
		if i > 0 {
			w.WriteString(" ")
		}
		i++
		w.WriteString(k)
		w.WriteString("=")
		writeValue(&w, v)
	}
	return w.String()
}

func writeValue(w *strings.Builder, v any) {
	switch v := v.(type) {
	case string:
		w.WriteString(v)
	case float64:
		w.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
	case int:
		w.WriteString(strconv.Itoa(v))
	case bool:
		w.WriteString(strconv.FormatBool(v))
	case []any:
		for i, v := range v {
			if i > 0 {
				w.WriteString(",")
			}
			writeValue(w, v)
		}
	default:
		panic("unknown label type")
	}
}
