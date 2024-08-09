package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestLinter(t *testing.T) {
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	cmd := exec.Command("lint-commit-or-rollback", ".")
	cmd.Dir = "testdata"
	output, err := cmd.CombinedOutput()
	assert.Error(t, err)
	expected := `
` + pwd + `/testdata/main.go:35:29: defer tx.CommitOrRollback(&err) should be deferred with a reference to a named error return parameter, but the function at /Users/alec/dev/pfi/cmd/lint-commit-or-rollback/testdata/main.go:29:6 has no named return parameters
` + pwd + `/testdata/main.go:44:28: defer tx.CommitOrRollback(&err) should be deferred with a reference to a named error return parameter, but the function at /Users/alec/dev/pfi/cmd/lint-commit-or-rollback/testdata/main.go:28:1 has no named return parameters
` + pwd + `/testdata/main.go:55:29: defer tx.CommitOrRollback(&err) should be deferred with a reference to a named error return parameter, but the function at /Users/alec/dev/pfi/cmd/lint-commit-or-rollback/testdata/main.go:49:6 has no named return parameters
	`
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(string(output)))
}
