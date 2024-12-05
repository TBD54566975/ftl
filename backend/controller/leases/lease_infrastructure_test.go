//go:build infrastructure

package leases_test

import (
	"testing"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestLeaseOnKube(t *testing.T) {
	tests := []in.ActionOrOption{in.WithKubernetes()}
	tests = append(tests, setupLeaseTests()...)
	in.Run(t,
		tests...,
	)
}
