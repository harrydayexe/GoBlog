// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package parser

import (
	"image"
	"io/fs"
	"log/slog"
	"sync"

	// Register the decoders whose headers image.DecodeConfig can read. SVG,
	// AVIF and WebP are not decodable here, so images in those formats are
	// rendered without dimensions.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// imageDimensions holds the intrinsic pixel size of an image asset. The zero
// value means the asset could not be measured.
type imageDimensions struct {
	width  int
	height int
}

// known reports whether both dimensions were read successfully.
func (d imageDimensions) known() bool {
	return d.width > 0 && d.height > 0
}

// imageMeasurer reads the intrinsic pixel dimensions of image assets from the
// assets filesystem, so rendered <img> tags can carry width and height and the
// browser can reserve space before the image loads.
//
// Only the image header is read, via [image.DecodeConfig]. Results are cached
// per path, including failures, so an image referenced from several posts is
// only read once.
//
// Measuring is always best-effort and never fails a parse: a nil measurer, no
// configured filesystem, a missing or unreadable file, and a format whose
// header cannot be decoded (SVG, AVIF, WebP) all mean "no dimensions".
type imageMeasurer struct {
	fsys   fs.FS
	logger *slog.Logger

	mu    sync.Mutex
	cache map[string]imageDimensions
}

// measure returns the dimensions of the asset at path, which must be a path
// within the assets filesystem, and whether they could be read.
func (m *imageMeasurer) measure(path string) (imageDimensions, bool) {
	if m == nil || m.fsys == nil || !fs.ValidPath(path) {
		return imageDimensions{}, false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if dim, cached := m.cache[path]; cached {
		return dim, dim.known()
	}

	dim := m.read(path)
	if m.cache == nil {
		m.cache = make(map[string]imageDimensions)
	}
	m.cache[path] = dim
	return dim, dim.known()
}

// read decodes the header of the asset at path. Failures are logged at debug
// level and reported as the zero imageDimensions.
func (m *imageMeasurer) read(path string) imageDimensions {
	f, err := m.fsys.Open(path)
	if err != nil {
		m.logger.Debug("image could not be opened, rendered without dimensions", slog.String("path", path), slog.Any("error", err))
		return imageDimensions{}
	}
	defer func() { _ = f.Close() }()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		m.logger.Debug("image could not be decoded, rendered without dimensions", slog.String("path", path), slog.Any("error", err))
		return imageDimensions{}
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return imageDimensions{}
	}

	return imageDimensions{width: cfg.Width, height: cfg.Height}
}
