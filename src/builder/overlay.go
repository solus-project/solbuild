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

package builder

import (
	// We'll end up using this later
	_ "github.com/Sirupsen/logrus"
	"path/filepath"
)

const (
	// OverlayRootDir is the root in which we form all solbuild cache paths,
	// these are the temp build roots that we happily throw away.
	OverlayRootDir = "/var/cache/solbuild"
)

// An Overlay is formed from a backing image & Package combination.
// Using this Overlay we can bring up new temporary build roots using the
// overlayfs kernel module.
type Overlay struct {
	Back    *BackingImage
	Package *Package

	WorkDir    string // WorkDir is the overlayfs workdir lock
	UpperDir   string // UpperDir is where real inode changes happen (transient)
	ImgDir     string // Where the profile is mounted (ro)
	MountPoint string // The actual mount point for the union'd directories
}

// NewOverlay creates a new Overlay for us in builds, etc.
//
// Unlike evobuild, we use fixed names within the more dynamic profile name,
// as opposed to a single dir with "unstable-x86_64" inside it, etc.
func NewOverlay(back *BackingImage, pkg *Package) *Overlay {
	// Ideally we could make this better..
	dirname := pkg.Name
	// i.e. /var/cache/solbuild/unstable-x86_64/nano
	basedir := filepath.Join(OverlayRootDir, back.Name, dirname)
	return &Overlay{
		Back:       back,
		Package:    pkg,
		WorkDir:    filepath.Join(basedir, "work"),
		UpperDir:   filepath.Join(basedir, "transient"),
		ImgDir:     filepath.Join(basedir, "image"),
		MountPoint: filepath.Join(basedir, "union"),
	}
}
