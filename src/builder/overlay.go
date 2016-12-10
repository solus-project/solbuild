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
	"github.com/solus-project/libosdev/commands"
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
	LockPath   string // Path to the lockfile for this overlay

	EnableTmpfs bool   // Whether to use tmpfs for the upperdir or not
	TmpfsSize   string // Size of the tmpfs to pass to mount, string form

	ExtraMounts []string // Any extra mounts to take care of when cleaning up

	mountedImg     bool // Whether we mounted the image or not
	mountedOverlay bool // Whether we mounted the overlay or not
	mountedVFS     bool // Whether we mounted vfs or not
	mountedTmpfs   bool // Whether we mounted tmpfs or not
}

// NewOverlay creates a new Overlay for us in builds, etc.
//
// Unlike evobuild, we use fixed names within the more dynamic profile name,
// as opposed to a single dir with "unstable-x86_64" inside it, etc.
func NewOverlay(profile *Profile, back *BackingImage, pkg *Package) *Overlay {
	// Ideally we could make this better..
	dirname := pkg.Name
	// i.e. /var/cache/solbuild/unstable-x86_64/nano
	basedir := filepath.Join(OverlayRootDir, profile.Name, dirname)
	return &Overlay{
		Back:           back,
		Package:        pkg,
		BaseDir:        basedir,
		WorkDir:        filepath.Join(basedir, "work"),
		UpperDir:       filepath.Join(basedir, "tmp"),
		ImgDir:         filepath.Join(basedir, "img"),
		MountPoint:     filepath.Join(basedir, "union"),
		LockPath:       fmt.Sprintf("%s.lock", basedir),
		mountedImg:     false,
		mountedOverlay: false,
		mountedVFS:     false,
		EnableTmpfs:    false,
		TmpfsSize:      "",
		mountedTmpfs:   false,
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
	log.Debug("Mounting overlayfs")

	mountMan := disk.GetMountManager()

	// Mount tmpfs as the root of all other mounts if requested
	if o.EnableTmpfs {
		if err := os.MkdirAll(o.BaseDir, 00755); err != nil {
			log.WithFields(log.Fields{
				"dir":   o.BaseDir,
				"error": err,
			}).Error("Failed to create tmpfs directory")
			return nil
		}

		log.WithFields(log.Fields{
			"point": o.BaseDir,
			"size":  o.TmpfsSize,
		}).Debug("Mounting root tmpfs")

		var tmpfsOptions []string
		if o.TmpfsSize != "" {
			tmpfsOptions = append(tmpfsOptions, fmt.Sprintf("size=%s", o.TmpfsSize))
		}
		tmpfsOptions = append(tmpfsOptions, []string{
			"rw",
			"relatime",
		}...)
		if err := mountMan.Mount("tmpfs-root", o.BaseDir, "tmpfs", tmpfsOptions...); err != nil {
			log.WithFields(log.Fields{
				"point": o.BaseDir,
				"size":  o.TmpfsSize,
			}).Error("Failed to mount root tmpfs")
			return err
		}
	}

	// Set up environment
	if err := o.EnsureDirs(); err != nil {
		return err
	}

	// First up, mount the backing image
	log.WithFields(log.Fields{
		"point": o.Back.ImagePath,
	}).Debug("Mounting backing image")
	if err := mountMan.Mount(o.Back.ImagePath, o.ImgDir, "auto", "ro", "loop"); err != nil {
		log.WithFields(log.Fields{
			"point": o.Back.ImagePath,
			"error": err,
		}).Error("Failed to mount backing image")
		return err
	}
	o.mountedImg = true

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

	// Must be done here before we do any more overlayfs work
	if err := EnsureEopkgLayout(o.MountPoint); err != nil {
		return err
	}

	return nil
}

// Unmount will tear down the overlay mount again
func (o *Overlay) Unmount() error {
	mountMan := disk.GetMountManager()

	for _, m := range o.ExtraMounts {
		mountMan.Unmount(m)
	}
	o.ExtraMounts = nil

	vfsPoints := []string{
		filepath.Join(o.MountPoint, "dev/pts"),
		filepath.Join(o.MountPoint, "dev/shm"),
		filepath.Join(o.MountPoint, "dev"),
		filepath.Join(o.MountPoint, "proc"),
		filepath.Join(o.MountPoint, "sys"),
	}
	if o.mountedVFS {
		for _, p := range vfsPoints {
			mountMan.Unmount(p)
		}
		o.mountedVFS = false
	}

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
	if o.mountedTmpfs {
		if err := mountMan.Unmount(o.UpperDir); err != nil {
			return err
		}
		o.mountedTmpfs = false
	}
	return nil
}

// MountVFS will bring up virtual filesystems within the chroot
func (o *Overlay) MountVFS() error {
	mountMan := disk.GetMountManager()

	vfsPoints := []string{
		filepath.Join(o.MountPoint, "dev"),
		filepath.Join(o.MountPoint, "dev/pts"),
		filepath.Join(o.MountPoint, "proc"),
		filepath.Join(o.MountPoint, "sys"),
		filepath.Join(o.MountPoint, "dev/shm"),
	}

	for _, p := range vfsPoints {
		if PathExists(p) {
			continue
		}

		log.WithFields(log.Fields{
			"dir": p,
		}).Debug("Creating VFS directory")

		if err := os.MkdirAll(p, 00755); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to create VFS directory")
			return err
		}
	}

	// Bring up dev
	log.WithFields(log.Fields{
		"vfs": "/dev",
	}).Debug("Mounting vfs")
	if err := mountMan.Mount("devtmpfs", vfsPoints[0], "devtmpfs", "nosuid", "mode=755"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to mount /dev")
		return err
	}
	o.mountedVFS = true

	// Bring up dev/pts
	log.WithFields(log.Fields{
		"vfs": "/dev/pts",
	}).Debug("Mounting vfs")
	if err := mountMan.Mount("devpts", vfsPoints[1], "devpts", "gid=5", "mode=620", "nosuid", "noexec"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to mount /dev/pts")
		return err
	}

	// Bring up proc
	log.WithFields(log.Fields{
		"vfs": "/proc",
	}).Debug("Mounting vfs")
	if err := mountMan.Mount("proc", vfsPoints[2], "proc", "nosuid", "noexec"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to mount /proc")
		return err
	}

	// Bring up sys
	log.WithFields(log.Fields{
		"vfs": "/sys",
	}).Debug("Mounting vfs")
	if err := mountMan.Mount("sysfs", vfsPoints[3], "sysfs"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to mount /sys")
		return err
	}

	// Bring up shm
	log.WithFields(log.Fields{
		"vfs": "/dev/shm",
	}).Debug("Mounting vfs")
	if err := mountMan.Mount("tmpfs-shm", vfsPoints[4], "tmpfs"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to mount /sys")
		return err
	}
	return nil
}

// ConfigureNetworking will add a loopback interface to the container so
// that localhost networking will still work
func (o *Overlay) ConfigureNetworking() error {
	ipCommand := "ip link set lo up"
	log.Debug("Configuring container networking")
	if err := commands.ChrootExec(o.MountPoint, ipCommand); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to configure networking")
		return err
	}
	return nil
}
