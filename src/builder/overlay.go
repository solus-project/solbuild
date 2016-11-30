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
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/solus-project/libosdev/disk"
	"os"
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
	Back    *BackingImage // This will be mounted at $dir/image
	Package *Package      // The package we intend to interact with

	BaseDir    string // BaseDir is the base directory containing the root
	WorkDir    string // WorkDir is the overlayfs workdir lock
	UpperDir   string // UpperDir is where real inode changes happen (tmp)
	ImgDir     string // Where the profile is mounted (ro)
	MountPoint string // The actual mount point for the union'd directories

	mountedImg     bool // Whether we mounted the image or not
	mountedOverlay bool // Whether we mounted the overlay or not
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
		Back:           back,
		Package:        pkg,
		BaseDir:        basedir,
		WorkDir:        filepath.Join(basedir, "work"),
		UpperDir:       filepath.Join(basedir, "tmp"),
		ImgDir:         filepath.Join(basedir, "img"),
		MountPoint:     filepath.Join(basedir, "union"),
		mountedImg:     false,
		mountedOverlay: false,
	}
}

// EnsureDirs is a helper to make sure we have all directories in place
func (o *Overlay) EnsureDirs() error {
	paths := []string{
		o.BaseDir,
		o.WorkDir,
		o.UpperDir,
		o.ImgDir,
		o.MountPoint,
	}

	for _, p := range paths {
		if PathExists(p) {
			continue
		}
		log.WithFields(log.Fields{
			"dir": p,
		}).Debug("Creating overlay storage directory")
		if err := os.MkdirAll(p, 00755); err != nil {
			log.WithFields(log.Fields{
				"dir":   p,
				"error": err,
			}).Error("Failed to create overlay storage directory")
			return err
		}
	}
	return nil
}

// CleanExisting will purge an existing overlayfs configuration if it
// exists.
func (o *Overlay) CleanExisting() error {
	if !PathExists(o.BaseDir) {
		return nil
	}
	log.WithFields(log.Fields{
		"dir": o.BaseDir,
	}).Debug("Removing stale workspace")
	if err := os.RemoveAll(o.BaseDir); err != nil {
		log.WithFields(log.Fields{
			"dir":   o.BaseDir,
			"error": err,
		}).Error("Failed to remove stale workspace")
		return err
	}
	return nil
}

// Mount will set up the overlayfs structure with the lower/upper respected
// properly.
func (o *Overlay) Mount() error {
	log.Info("Mounting overlayfs")

	mountMan := disk.GetMountManager()

	// First up, mount the backing image
	log.WithFields(log.Fields{
		"point": o.Back.ImagePath,
	}).Debug("Mounting backing image")
	if err := mountMan.Mount(o.Back.ImagePath, o.ImgDir, "auto", "loop", "ro"); err != nil {
		log.WithFields(log.Fields{
			"point": o.Back.ImagePath,
			"error": err,
		}).Error("Failed to mount backing image")
		return err
	}
	o.mountedImg = true

	// TODO: Mount tmpfs at upperdir if requested!

	// Now mount the overlayfs
	log.WithFields(log.Fields{
		"upper":   o.UpperDir,
		"lower":   o.ImgDir,
		"workdir": o.WorkDir,
		"target":  o.MountPoint,
	}).Debug("Mounting overlayfs")

	// Mounting overlayfs..
	err := mountMan.Mount("overlay", o.MountPoint, "overlay",
		fmt.Sprintf("lowerdir=%s", o.ImgDir),
		fmt.Sprintf("upperdir=%s", o.UpperDir),
		fmt.Sprintf("workdir=%s", o.WorkDir))

	// Check non-fatal..
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"point": o.MountPoint,
		}).Error("Failed to mount overlayfs")
		return err
	}
	o.mountedOverlay = true
	return nil
}

// Unmount will tear down the overlay mount again
func (o *Overlay) Unmount() error {
	mountMan := disk.GetMountManager()

	if o.mountedImg {
		if err := mountMan.Unmount(o.ImgDir); err != nil {
			return err
		}
		o.mountedImg = false
	}
	if o.mountedOverlay {
		if err := mountMan.Unmount(o.MountPoint); err != nil {
			return err
		}
		o.mountedOverlay = false
	}
	return nil
}
