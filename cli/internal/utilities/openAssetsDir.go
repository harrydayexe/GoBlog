// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package utilities

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// DefaultAssetsDirName is the assets directory used, relative to the posts
// directory, when no --assets-dir flag is given.
const DefaultAssetsDirName = "images"

// OpenAssetsDir opens the assets directory as an [os.Root], so files served
// or copied from it cannot escape the directory via ".." or symbolic links.
//
// When assetsDir is empty it defaults to <postsDir>/images. If the resolved
// directory does not exist, OpenAssetsDir returns a nil Root and nil error:
// asset support is silently off. Any other failure, such as the path not
// being a directory, is returned as an error.
//
// The caller must Close a non-nil Root.
func OpenAssetsDir(assetsDir, postsDir string) (*os.Root, error) {
	if assetsDir == "" {
		assetsDir = filepath.Join(postsDir, DefaultAssetsDirName)
	}

	root, err := os.OpenRoot(filepath.Clean(assetsDir))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return root, err
}
