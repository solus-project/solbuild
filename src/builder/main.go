//
// Copyright © 2016 Ikey Doherty <ikey@solus-project.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Package builder provides all the solbuild specific functionality
package builder

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// ImagesDir is where we keep the rootfs images for build profiles
	ImagesDir = "/var/lib/solbuild/images"

	// ImageSuffix is the common suffix for all solbuild images
	ImageSuffix = ".img"

	// ImageCompressedSuffix is the common suffix for a fetched evobuild image
	ImageCompressedSuffix = ".img.xz"

	// ImageBaseURI is the storage area for base images
	ImageBaseURI = "https://solus-project.com/image_root"
)

// PathExists is a helper function to determine the existence of a file path
func PathExists(path string) bool {
	if st, err := os.Stat(path); err == nil && st != nil {
		return true
	}
	return false
}

// A BackingImage is the core of any given profile
type BackingImage struct {
	Name        string // Name of the profile
	ImagePath   string // Absolute path to the .img file
	ImagePathXZ string // Absolute path to the .img.xz file
	ImageURI    string // URI of the image origin
}

// IsInstalled will determine whether the given backing image has been installed
// to the global image directory or not.
func (b *BackingImage) IsInstalled() bool {
	return PathExists(b.ImagePath)
}

// IsFetched will determine whether or not the XZ image itself has been fetched
func (b *BackingImage) IsFetched() bool {
	return PathExists(b.ImagePathXZ)
}

// NewBackingImage will return a correctly configured backing image for
// usage.
func NewBackingImage(name string) *BackingImage {
	return &BackingImage{
		Name:        name,
		ImagePath:   filepath.Join(ImagesDir, name+ImageSuffix),
		ImagePathXZ: filepath.Join(ImagesDir, name+ImageCompressedSuffix),
		ImageURI:    fmt.Sprintf("%s/%s%s", ImageBaseURI, name, ImageCompressedSuffix),
	}
}
