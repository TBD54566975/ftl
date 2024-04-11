package ftl

import (
	"fmt"
	"regexp"

	"golang.org/x/mod/semver"
)

// IsRelease returns true if the version is a release version.
func IsRelease(v string) bool {
	return regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(v)
}

// IsVersionAtLeastMin returns true if any of the following are true:
//   - minVersion is not defined (i.e. is emptystring)
//   - v or minVersion is not a release version
//   - v > minVersion when both v and minVersion are release versions
func IsVersionAtLeastMin(v string, minVersion string) bool {
	if minVersion == "" {
		return true
	}
	if !IsRelease(v) || !IsRelease(minVersion) {
		return true
	}
	return semver.Compare("v"+v, "v"+minVersion) >= 0
}

type VersionNotSupportedError struct {
	FTLVersion    string
	MinFTLVersion string
}

func (e *VersionNotSupportedError) Error() string {
	return fmt.Sprintf("FTL version '%v' predates the minimum version '%v'", e.FTLVersion, e.MinFTLVersion)
}

// Version of FTL binary (set by linker).
var Version = "dev"

// Timestamp of FTL binary (set by linker).
var Timestamp = "0"
