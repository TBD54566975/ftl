package ftl

import "regexp"

// IsRelease returns true if the version is a release version.
func IsRelease(v string) bool {
	return regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(v)
}

// Version of FTL binary (set by linker).
var Version = "dev"

// Timestamp of FTL binary (set by linker).
var Timestamp = "0"
