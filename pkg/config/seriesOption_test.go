// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"testing"
	"testing/fstest"
)

// TestSeriesFile_Enabled asserts that series are only on when both halves of the
// location are present: a filesystem without a path names nothing, and a path
// without a filesystem cannot be read.
func TestSeriesFile_Enabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		file SeriesFile
		want bool
	}{
		{name: "zero value", file: SeriesFile{}, want: false},
		{name: "filesystem only", file: SeriesFile{FS: fstest.MapFS{}}, want: false},
		{name: "path only", file: SeriesFile{Path: "series.yml"}, want: false},
		{name: "both", file: SeriesFile{FS: fstest.MapFS{}, Path: "series.yml"}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.file.Enabled(); got != tt.want {
				t.Errorf("Enabled() = %t, want %t", got, tt.want)
			}
		})
	}
}

// TestWithSeriesFile_RoundTrip asserts a configured series file survives being
// converted back into an option, which is how the server forwards it on.
func TestWithSeriesFile_RoundTrip(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{}

	var applied SeriesFile
	WithSeriesFile(fsys, "series.yml").WithSeriesFileFunc(&applied)

	var round SeriesFile
	applied.AsOption().WithSeriesFileFunc(&round)

	if round.Path != "series.yml" || round.FS == nil {
		t.Errorf("AsOption() round trip = %+v, want the original filesystem and path", round)
	}
}
