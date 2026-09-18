// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config_test

import (
	"log/slog"
	"testing"
	"testing/fstest"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
)

func TestAsGeneratorOption_LiftsLogger(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	opt := config.WithLogger(logger).AsGeneratorOption()

	if opt.WithLoggerFunc == nil {
		t.Fatal("AsGeneratorOption() did not carry the base option's WithLoggerFunc")
	}
	var got config.Logger
	opt.WithLoggerFunc(&got)
	if got.Logger != logger {
		t.Errorf("Logger = %v, want %v", got.Logger, logger)
	}
}

func TestAsGeneratorOption_LiftsBlogRoot(t *testing.T) {
	t.Parallel()

	opt := config.WithBlogRoot("/blog/").AsGeneratorOption()

	if opt.WithBlogRootFunc == nil {
		t.Fatal("AsGeneratorOption() did not carry the base option's WithBlogRootFunc")
	}
	var got config.BlogRoot
	opt.WithBlogRootFunc(&got)
	if got != "/blog/" {
		t.Errorf("BlogRoot = %q, want %q", got, "/blog/")
	}
}

func TestAsGeneratorOption_LiftsAssetsDir(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{"cat.png": {Data: []byte("png")}}
	opt := config.WithAssetsDir(fsys).AsGeneratorOption()

	if opt.WithAssetsDirFunc == nil {
		t.Fatal("AsGeneratorOption() did not carry the base option's WithAssetsDirFunc")
	}
	var got config.AssetsDir
	opt.WithAssetsDirFunc(&got)
	if !got.Enabled() {
		t.Error("AssetsDir.Enabled() = false, want true")
	}
}

// AsGeneratorOption only lifts the BaseOption: the generator-specific function
// pointers must stay nil so the option-application chain does not pick a
// generator field the caller never set.
func TestAsGeneratorOption_LeavesGeneratorFieldsUnset(t *testing.T) {
	t.Parallel()

	opt := config.WithLogger(slog.New(slog.DiscardHandler)).AsGeneratorOption()

	if opt.WithRawOutputFunc != nil {
		t.Error("WithRawOutputFunc is set, want nil")
	}
	if opt.WithSiteTitleFunc != nil {
		t.Error("WithSiteTitleFunc is set, want nil")
	}
	if opt.WithCustomDataFunc != nil {
		t.Error("WithCustomDataFunc is set, want nil")
	}
}
