//go:build integration

package leases_test

import (
	"testing"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestLease(t *testing.T) {
	in.Run(t,
		setupLeaseTests()...,
	)
}
