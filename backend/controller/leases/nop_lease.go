package leases

// NoopLease is a no-op implementation of [Lease].
var NoopLease Lease = noopLease{}

type noopLease struct{}

func (noopLease) Release() error { return nil }
