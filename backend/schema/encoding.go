package schema

import (
	"fmt"
	"strings"
)

// This file contains the unmarshalling logic as well as support methods for
// visiting and type safety.

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "  " + strings.Join(strings.Split(s, "\n"), "\n  ")
}

func encodeMetadata(metadata []Metadata) string {
	if len(metadata) == 0 {
		return ""
	}
	w := &strings.Builder{}
	fmt.Fprintln(w)
	for _, c := range metadata {
		fmt.Fprint(w, indent(c.String()))
	}
	return w.String()
}

func encodeComments(comments []string) string {
	if len(comments) == 0 {
		return ""
	}
	w := &strings.Builder{}
	for _, c := range comments {
		fmt.Fprintf(w, "// %s\n", c)
	}
	return w.String()
}

func makeRef(module, name string) string {
	if module == "" {
		return name
	}
	return module + "." + name
}
