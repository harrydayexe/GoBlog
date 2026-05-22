// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import "runtime/debug"

// buildVersion returns the version string for the running binary.
// When GoReleaser ldflags are present the injected value is returned as-is.
// Otherwise the value is derived from the embedded Go module build metadata,
// so binaries installed via go install report the module tag. It is safe to
// call from multiple goroutines.
func buildVersion() string {
	if version != "dev" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version
	}

	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}

	return version
}
