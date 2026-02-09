package changelog

import (
	"fmt"
	"strconv"
	"strings"
)

// BumpType represents the kind of semver bump to apply.
type BumpType int

const (
	BumpPatch BumpType = iota
	BumpMinor
	BumpMajor
)

// Version represents a parsed semantic version.
type Version struct {
	Major int
	Minor int
	Patch int
}

// ParseVersion parses a semver tag like "v1.2.3" or "1.2.3".
func ParseVersion(tag string) (Version, error) {
	s := strings.TrimPrefix(tag, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("could not parse tag %q", tag)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("could not parse tag %q", tag)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("could not parse tag %q", tag)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("could not parse tag %q", tag)
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// String returns the version formatted as "v1.2.3".
func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Bump returns a new Version with the given bump type applied.
func (v Version) Bump(bt BumpType) Version {
	switch bt {
	case BumpMajor:
		return Version{Major: v.Major + 1}
	case BumpMinor:
		return Version{Major: v.Major, Minor: v.Minor + 1}
	default:
		return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
	}
}

// DetermineBump scans conventional commit subjects and returns the appropriate bump type.
func DetermineBump(subjects []string) BumpType {
	bump := BumpPatch

	for _, s := range subjects {
		lower := strings.ToLower(s)

		// Check for breaking changes
		if strings.Contains(s, "!:") || strings.Contains(lower, "breaking change") {
			return BumpMajor
		}

		// Check for feat prefix
		if strings.HasPrefix(lower, "feat") {
			bump = BumpMinor
		}
	}

	return bump
}

// BumpLabel returns a human-readable label for the bump type.
func BumpLabel(bt BumpType) string {
	switch bt {
	case BumpMajor:
		return "major — breaking changes detected"
	case BumpMinor:
		return "minor — new features detected"
	default:
		return "patch"
	}
}
