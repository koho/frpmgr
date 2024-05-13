package version

import (
	"fmt"
	"regexp"
	"strconv"
)

// Parse parses a version string into its major, minor, and patch components.
func Parse(s string) (major, minor, patch int, err error) {
	re := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		err = fmt.Errorf("invalid version string: %s", s)
		return
	}
	major, err = strconv.Atoi(matches[1])
	if err != nil {
		return
	}
	minor, err = strconv.Atoi(matches[2])
	if err != nil {
		return
	}
	patch, err = strconv.Atoi(matches[3])
	return
}
