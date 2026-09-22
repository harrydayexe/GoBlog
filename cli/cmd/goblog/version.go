// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import "runtime/debug"

// buildVersion returns the version string for the running binary.
// Released binaries carry the version injected by GoReleaser's ldflags, which
// is returned as-is. The CLI module is never published to the module proxy, so
// a locally built binary has no module tag to fall back on; the embedded Go
// module build metadata is still consulted in case one is present, and
// otherwise the version reads "dev". It is safe to call from multiple
// goroutines.
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
