package model

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestDeploymentKey(t *testing.T) {
	expected := NewDeploymentKey()
	assert.True(t, strings.HasPrefix(expected.String(), "D"))
	actual, err := ParseDeploymentKey(expected.String())
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestRunnerKey(t *testing.T) {
	expected := NewRunnerKey()
	assert.True(t, strings.HasPrefix(expected.String(), "R"))
	actual, err := ParseRunnerKey(expected.String())
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
