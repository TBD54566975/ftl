package ftl

import (
	"regexp"
	"strconv"
)

const VersionNumberParts int = 3

// IsRelease returns true if the version is a release version.
func IsRelease(v string) bool {
	return regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(v)
}

// IsVersionAtLeastMin returns true if either v or minVersion is not a release version, or if v > minVersion when both v and minVersion are release versions
func IsVersionAtLeastMin(v string, minVersion string) (bool, error) {
	if !IsRelease(v) || !IsRelease(minVersion) {
		return true, nil
	}
	vParsed := regexp.MustCompile(`\d+`).FindAllString(v, VersionNumberParts)
	minVParsed := regexp.MustCompile(`\d+`).FindAllString(minVersion, VersionNumberParts)
	for i := range VersionNumberParts {
		vInt, err := strconv.Atoi(vParsed[i])
		if err != nil {
			return false, err
		}
		minVInt, err := strconv.Atoi(minVParsed[i])
		if err != nil {
			return false, err
		}
		if vInt < minVInt {
			return false, nil
		}
	}
	return true, nil
}

// Version of FTL binary (set by linker).
var Version = "dev"

// Timestamp of FTL binary (set by linker).
var Timestamp = "0"
