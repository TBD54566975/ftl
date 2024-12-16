//go:build integration

package leases_test

import (
	"testing"

	in "github.com/block/ftl/internal/integration"
)

func TestLease(t *testing.T) {
	in.Run(t,
		setupLeaseTests()...,
	)
}
