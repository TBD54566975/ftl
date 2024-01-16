package ftl

import "strings"

// IsRelease returns true if the version is a release version.
func IsRelease(v string) bool {
	return v != "dev" && !strings.HasSuffix(v, "-dirty")
}

// Version of FTL binary (set by linker).
var Version = "dev"

// Timestamp of FTL binary (set by linker).
var Timestamp = "0"
