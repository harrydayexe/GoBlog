package main

import (
	"regexp"
	"runtime/debug"
)

// pseudoVersionRe matches Go pseudo-versions: v0.0.0-YYYYMMDDhhmmss-abcdefabcdef
var pseudoVersionRe = regexp.MustCompile(`-(\d{14})-([0-9a-f]{12})$`)

// buildVersion returns the version, commit, and date for the running binary.
// When GoReleaser ldflags are present they are returned as-is. Otherwise the
// values are derived from the embedded Go module and VCS build metadata, so
// binaries installed via go install report useful information rather than the
// placeholder defaults. It is safe to call from multiple goroutines.
func buildVersion() (ver, com, dat string) {
	// Ldflag path: GoReleaser has already injected real values.
	if version != "dev" {
		return version, commit, date
	}

	ver, com, dat = version, commit, date

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	if v := info.Main.Version; v != "" && v != "(devel)" {
		ver = v
	}

	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if len(s.Value) >= 7 {
				com = s.Value[:7]
			} else {
				com = s.Value
			}
		case "vcs.time":
			dat = s.Value
		case "vcs.modified":
			if s.Value == "true" && com != "none" {
				com += "+dirty"
			}
		}
	}

	// For go install builds, vcs.* settings are not populated but the version
	// is a pseudo-version that encodes the date and commit SHA.
	if com == "none" && dat == "unknown" {
		if m := pseudoVersionRe.FindStringSubmatch(ver); m != nil {
			raw := m[1] // YYYYMMDDhhmmss
			dat = raw[:4] + "-" + raw[4:6] + "-" + raw[6:8] + "T" + raw[8:10] + ":" + raw[10:12] + ":" + raw[12:14] + "Z"
			com = m[2][:7]
		}
	}

	return
}
