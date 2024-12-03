package schema

// Runtime is data that is populated at runtime by an FTL subsystem.
type Runtime interface {
	runtime()
}
