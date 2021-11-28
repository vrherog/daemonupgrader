package version

import (
	"strconv"
	"strings"
)

var (
	BuildVersion string
)

type Version struct {
	Major    int
	Minor    int
	Build    int
	Revision string
}

func ParseVersion(version string) (result Version, ok bool) {
	var parts = strings.Split(version, `.`)
	if len(parts) > 4 {
		ok = false
		return
	}
	if len(parts) > 3 {
		result.Revision = parts[3]
		ok = true
	}
	if len(parts) > 2 {
		var build, err = strconv.Atoi(parts[2])
		ok = err == nil
		if ok {
			result.Build = build
		}
	}
	if ok && len(parts) > 1 {
		var minor, err = strconv.Atoi(parts[1])
		ok = err == nil
		if ok {
			result.Minor = minor
		}
	}
	if ok && len(parts) > 0 {
		var major, err = strconv.Atoi(parts[0])
		ok = err == nil
		if ok {
			result.Major = major
		}
	}
	return
}

func (v Version) Compare(other Version) int8 {
	if v.Major > other.Major {
		return 1
	} else if v.Major < other.Major {
		return -1
	}
	if v.Minor > other.Minor {
		return 1
	} else if v.Minor < other.Minor {
		return -1
	}
	if v.Build > other.Build {
		return 1
	} else if v.Build < other.Build {
		return -1
	}
	if v.Revision > other.Revision {
		return 1
	} else if v.Revision < other.Revision {
		return -1
	}
	return 0
}

func CompareVersion(v1 string, v2 string) (int8, bool) {
	if ver1, ok := ParseVersion(v1); ok {
		if ver2, ok := ParseVersion(v2); ok {
			return ver1.Compare(ver2), true
		}
	}
	return 0, false
}
