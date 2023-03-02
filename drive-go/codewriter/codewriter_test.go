package codewriter_test

import (
	"testing"

	"github.com/TBD54566975/ftl/drive-go/codewriter"
	"github.com/alecthomas/assert/v2"
)

func TestCodeWriter(t *testing.T) {
	w := codewriter.New("pkg")
	w.L("func hello() {")
	w.In(func(w *codewriter.Writer) {
		w.L(`println("hello")`)
	})
	w.L("}")
	expected := `func hello() {
  println("hello")
}
`
	assert.Equal(t, expected, w.Body())
}
