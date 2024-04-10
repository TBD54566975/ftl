package ftl

import (
	"regexp"

	"golang.org/x/mod/semver"
)

const VersionNumberParts int = 3

// IsRelease returns true if the version is a release version.
func IsRelease(v string) bool {
	return regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(v)
}

// IsVersionAtLeastMin returns true if any of the following are true:
//   - minVersion is not defined (i.e. is emptystring)
//   - v or minVersion is not a release version
//   - v > minVersion when both v and minVersion are release versions
func IsVersionAtLeastMin(v string, minVersion string) (bool, error) {
	if minVersion == "" {
		return true, nil
	}
	if !IsRelease(v) || !IsRelease(minVersion) {
		return true, nil
	}
	return semver.Compare("v"+v, "v"+minVersion) >= 0, nil
}

// VersionIsMock is set by tests and used to block evaluation of versions that look like release versions but are not real.
var VersionIsMock = false

// Version of FTL binary (set by linker).
var Version = "dev"

// Timestamp of FTL binary (set by linker).
var Timestamp = "0"
