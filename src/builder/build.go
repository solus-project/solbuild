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
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/solus-project/libosdev/disk"
	"os"
	"path/filepath"
)

// CreateDirs creates any directories we may need later on
func (p *Package) CreateDirs(o *Overlay) error {
	dirs := []string{
		p.GetWorkDir(o),
		p.GetSourceDir(o),
		p.GetCcacheDir(o),
	}
	for _, p := range dirs {
		if err := os.MkdirAll(p, 00755); err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"dir":   p,
			}).Error("Failed to create required directory")
			return err
		}
	}

	// Fix up the ccache directories
	if p.Type == PackageTypeXML {
		// Ensure we have root owned ccache
		if err := os.MkdirAll(LegacyCcacheDirectory, 00755); err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"dir":   p,
			}).Error("Failed to create ccache directory")
			return err
		}
	} else {
		// Ensure we have root owned ccache
		if err := os.MkdirAll(CcacheDirectory, 00755); err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"dir":   p,
			}).Error("Failed to create ccache directory")
			return err
		}
		if err := os.Chown(CcacheDirectory, BuildUserID, BuildUserGID); err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"dir":   p,
			}).Error("Failed to chown ccache directory")
			return err
		}
	}

	return nil
}

// FetchSources will attempt to fetch the sources from the network
// if necessary
func (p *Package) FetchSources(o *Overlay) error {
	for _, source := range p.Sources {
		// Already fetched, skip it
		if source.IsFetched() {
			continue
		}
		if err := source.Fetch(); err != nil {
			log.WithFields(log.Fields{
				"error":  err,
				"source": source.GetIdentifier(),
			}).Error("Failed to fetch source")
			return err
		}
	}
	return nil
}

// BindSources will make the sources available to the chroot by bind mounting
// them into place.
func (p *Package) BindSources(o *Overlay) error {
	mountMan := disk.GetMountManager()

	for _, source := range p.Sources {
		sourceDir := p.GetSourceDir(o)
		bindConfig := source.GetBindConfiguration(sourceDir)

		// Ensure sources tree exists
		if !PathExists(sourceDir) {
			if err := os.MkdirAll(sourceDir, 00755); err != nil {
				log.WithFields(log.Fields{
					"dir":   sourceDir,
					"error": err,
				}).Error("Failed to create source directory")
				return err
			}
		}

		// Find the target path in the chroot
		log.WithFields(log.Fields{
			"target": bindConfig.BindTarget,
		}).Debug("Exposing source to container")

		if st, err := os.Stat(bindConfig.BindSource); err == nil && st != nil {
			if st.IsDir() {
				if err := os.MkdirAll(bindConfig.BindTarget, 00755); err != nil {
					log.WithFields(log.Fields{
						"target": bindConfig.BindTarget,
						"error":  err,
					}).Error("Failed to create bind mount target")
					return nil
				}
			} else {
				if err := TouchFile(bindConfig.BindTarget); err != nil {
					log.WithFields(log.Fields{
						"target": bindConfig.BindTarget,
						"error":  err,
					}).Error("Failed to create bind mount target")
					return nil
				}
			}
		}

		// Bind mount local source into chroot
		if err := mountMan.BindMount(bindConfig.BindSource, bindConfig.BindTarget, "ro"); err != nil {
			log.WithFields(log.Fields{
				"target": bindConfig.BindTarget,
				"error":  err,
			}).Error("Failed to bind mount source")
			return err
		}

		// Account for these to help cleanups
		o.ExtraMounts = append(o.ExtraMounts, bindConfig.BindTarget)
	}
	return nil
}

// BindCcache will make the ccache directory available to the build
func (p *Package) BindCcache(o *Overlay) error {
	mountMan := disk.GetMountManager()
	ccacheDir := p.GetCcacheDir(o)

	var ccacheSource string
	if p.Type == PackageTypeXML {
		ccacheSource = LegacyCcacheDirectory
	} else {
		ccacheSource = CcacheDirectory
	}

	log.WithFields(log.Fields{
		"dir": ccacheDir,
	}).Debug("Exposing ccache to build")

	// Bind mount local ccache into chroot
	if err := mountMan.BindMount(ccacheSource, ccacheDir); err != nil {
		log.WithFields(log.Fields{
			"target": ccacheDir,
			"error":  err,
		}).Error("Failed to bind mount ccache")
		return err
	}
	o.ExtraMounts = append(o.ExtraMounts, ccacheDir)
	return nil
}

// GetWorkDir will return the externally visible work directory for the
// given build type.
func (p *Package) GetWorkDir(o *Overlay) string {
	return filepath.Join(o.MountPoint, p.GetWorkDirInternal()[1:])
}

// GetWorkDirInternal returns the internal chroot path for the work directory
func (p *Package) GetWorkDirInternal() string {
	if p.Type == PackageTypeXML {
		return "/WORK"
	}
	return filepath.Join(BuildUserHome, "work")
}

// GetSourceDir will return the externally visible work directory
func (p *Package) GetSourceDir(o *Overlay) string {
	return filepath.Join(o.MountPoint, p.GetSourceDirInternal()[1:])
}

// GetSourceDirInternal will return the chroot-internal source directory
// for the given build type.
func (p *Package) GetSourceDirInternal() string {
	if p.Type == PackageTypeXML {
		return "/var/cache/eopkg/archives"
	}
	return filepath.Join(BuildUserHome, "YPKG", "sources")
}

// GetCcacheDir will return the externally visible ccache directory
func (p *Package) GetCcacheDir(o *Overlay) string {
	return filepath.Join(o.MountPoint, p.GetCcacheDirInternal()[1:])
}

// GetCcacheDirInternal will return the chroot-internal ccache directory
// for the given build type
func (p *Package) GetCcacheDirInternal() string {
	if p.Type == PackageTypeXML {
		return "/root/.ccache"
	}
	return filepath.Join(BuildUserHome, ".ccache")
}

// CopyAssets will copy all of the required assets into the builder root
func (p *Package) CopyAssets(h *PackageHistory, o *Overlay) error {
	baseDir := filepath.Dir(p.Path)

	if abs, err := filepath.Abs(baseDir); err == nil {
		baseDir = abs
	} else {
		return err
	}

	copyPaths := []string{
		filepath.Base(p.Path),
		"files",
		"comar",
		"component.xml",
	}

	if p.Type == PackageTypeXML {
		copyPaths = append(copyPaths, "actions.py")
	}

	// This should be changed for ypkg.
	destdir := p.GetWorkDir(o)

	for _, pat := range copyPaths {
		fso := filepath.Join(baseDir, pat)
		newDest := destdir
		if p.Type == PackageTypeXML && pat == "component.xml" {
			newDest = filepath.Dir(destdir)
		}
		if err := CopyAll(fso, newDest); err != nil {
			return err
		}
	}

	if h == nil {
		return nil
	}
	// Write the history file out
	histPath := filepath.Join(destdir, "history.xml")
	return h.WriteXML(histPath)
}

// BuildYpkg will take care of the ypkg specific build process and is called only
// by Build()
func (p *Package) BuildYpkg(notif PidNotifier, usr *UserInfo, pman *EopkgManager, overlay *Overlay, h *PackageHistory) error {
	log.Debug("Writing packager file")
	fp := filepath.Join(overlay.MountPoint, BuildUserHome, ".solus", "packager")
	fpd := filepath.Dir(fp)

	if !PathExists(fpd) {
		if err := os.MkdirAll(fpd, 00755); err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"dir":   fpd,
			}).Error("Failed to create packager directory")
			return err
		}
	}

	if err := usr.WritePackager(fp); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  fp,
		}).Error("Failed to write packager file")
		return err
	}

	wdir := p.GetWorkDirInternal()
	ymlFile := filepath.Join(wdir, filepath.Base(p.Path))
	cmd := fmt.Sprintf("ypkg-install-deps -f %s", ymlFile)
	if DisableColors {
		cmd += " -n"
	}

	// Install build dependencies
	log.WithFields(log.Fields{
		"buildFile": ymlFile,
	}).Info("Installing build dependencies")

	if err := ChrootExec(notif, overlay.MountPoint, cmd); err != nil {
		log.WithFields(log.Fields{
			"buildFile": ymlFile,
			"error":     err,
		}).Error("Failed to install build dependencies")
		return err
	}
	notif.SetActivePID(0)

	// Cleanup now
	log.Debug("Stopping D-BUS")
	if err := pman.StopDBUS(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to stop d-bus")
		return err
	}

	// Chwn the directory before bringing up sources
	cmd = fmt.Sprintf("chown -R %s:%s %s", BuildUser, BuildUser, BuildUserHome)
	if err := ChrootExec(notif, overlay.MountPoint, cmd); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to set home directory permissions")
		return err
	}
	notif.SetActivePID(0)

	// Now kill networking
	if !p.CanNetwork {
		if err := DropNetworking(); err != nil {
			return err
		}

		// Ensure the overlay can network on localhost only
		if err := overlay.ConfigureNetworking(); err != nil {
			return err
		}
	} else {
		log.Warning("Package has explicitly requested networking, sandboxing disabled")
	}

	// Bring up sources
	if err := p.BindSources(overlay); err != nil {
		log.Error("Cannot continue without sources")
		return err
	}

	// Reaffirm the layout
	if err := EnsureEopkgLayout(overlay.MountPoint); err != nil {
		return err
	}

	// Ensure we have ccache available
	if err := p.BindCcache(overlay); err != nil {
		return err
	}

	// Now recopy the assets prior to build
	if err := pman.CopyAssets(); err != nil {
		return err
	}

	// Now build the package
	cmd = fmt.Sprintf("/bin/su %s -- fakeroot ypkg-build -D %s %s", BuildUser, wdir, ymlFile)
	if DisableColors {
		cmd += " -n"
	}
	// Pass unix timestamp of last git update
	if h != nil && len(h.Updates) > 0 {
		cmd += fmt.Sprintf(" -t %v", h.Updates[0].Time.UTC().Unix())
	}

	log.WithFields(log.Fields{
		"package": p.Name,
	}).Info("Now starting build of package")
	if err := ChrootExec(notif, overlay.MountPoint, cmd); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to build package")
		return err
	}
	notif.SetActivePID(0)
	return nil
}

// BuildXML will take care of building the legacy pspec.xml format, and is called only
// by Build()
func (p *Package) BuildXML(notif PidNotifier, pman *EopkgManager, overlay *Overlay) error {
	// Just straight up build it with eopkg
	log.Warning("Full sandboxing is not possible with legacy format")

	wdir := p.GetWorkDirInternal()
	xmlFile := filepath.Join(wdir, filepath.Base(p.Path))

	// Bring up sources
	if err := p.BindSources(overlay); err != nil {
		log.Error("Cannot continue without sources")
		return err
	}

	// Ensure we have ccache available
	if err := p.BindCcache(overlay); err != nil {
		return err
	}

	// Now recopy the assets prior to build
	if err := pman.CopyAssets(); err != nil {
		return err
	}

	// Now build the package, ignore-sandbox in case someone is stupid
	// and activates it in eopkg.conf..
	cmd := eopkgCommand(fmt.Sprintf("eopkg build --ignore-sandbox --yes-all -O %s %s", wdir, xmlFile))
	log.WithFields(log.Fields{
		"package": p.Name,
	}).Info("Now starting build of package")
	if err := ChrootExec(notif, overlay.MountPoint, cmd); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to build package")
		return err
	}
	notif.SetActivePID(0)

	// Now we can stop dbus..
	log.Debug("Stopping D-BUS")
	if err := pman.StopDBUS(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to stop d-bus")
		return err
	}
	notif.SetActivePID(0)
	return nil
}

// CollectAssets will search for the build files and copy them back to the
// users current directory. If solbuild was invoked via sudo, solbuild will
// then attempt to set the owner as the original user.
func (p *Package) CollectAssets(overlay *Overlay, usr *UserInfo) error {
	collectionDir := p.GetWorkDir(overlay)
	collections, _ := filepath.Glob(filepath.Join(collectionDir, "*.eopkg"))
	if len(collections) < 1 {
		log.Error("Mysterious lack of eopkg files is mysterious")
		return errors.New("Internal error: .eopkg files are missing")
	}

	if p.Type == PackageTypeYpkg {
		pspecs, _ := filepath.Glob(filepath.Join(collectionDir, "pspec_*.xml"))
		collections = append(collections, pspecs...)
	}

	log.WithFields(log.Fields{
		"numFiles": len(collections),
	}).Debug("Collecting files")

	for _, p := range collections {
		tgt, err := filepath.Abs(filepath.Join(".", filepath.Base(p)))
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Unable to find working directory!")
			return err
		}

		log.WithFields(log.Fields{
			"file": filepath.Base(p),
		}).Info("Collecting build artifact")

		if err := disk.CopyFile(p, tgt); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Unable to collect build file")
			return err
		}

		log.WithFields(log.Fields{
			"uid":  usr.UID,
			"gid":  usr.GID,
			"file": filepath.Base(p),
		}).Debug("Setting file ownership for current user")

		if err = os.Chown(tgt, usr.UID, usr.GID); err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"file":  filepath.Base(p),
			}).Error("Error in restoring file ownership")
		}
	}
	return nil
}

// Build will attempt to build the package in the overlayfs system
func (p *Package) Build(notif PidNotifier, history *PackageHistory, profile *Profile, pman *EopkgManager, overlay *Overlay) error {
	log.WithFields(log.Fields{
		"profile": overlay.Back.Name,
		"version": p.Version,
		"package": p.Name,
		"type":    p.Type,
		"release": p.Release,
	}).Info("Building package")

	usr := GetUserInfo()

	var env []string
	if p.Type == PackageTypeXML {
		env = SaneEnvironment("root", "/root")
	} else {
		env = SaneEnvironment(BuildUser, BuildUserHome)
	}
	ChrootEnvironment = env

	// Set up environment
	if err := overlay.CleanExisting(); err != nil {
		return err
	}

	// Bring up the root
	if err := p.ActivateRoot(overlay); err != nil {
		return err
	}

	// Ensure source assets are in place
	if err := p.CopyAssets(history, overlay); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to copy required source assets")
		return err
	}

	log.Info("Validating sources")
	if err := p.FetchSources(overlay); err != nil {
		return err
	}

	// Set up package manager
	if err := pman.Init(); err != nil {
		return err
	}

	// Bring up dbus to do Things
	log.Debug("Starting D-BUS")
	if err := pman.StartDBUS(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to start d-bus")
		return err
	}

	// Get the repos in place before asserting anything
	if err := p.ConfigureRepos(notif, overlay, pman, profile); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Configuring repositories failed")
		return err
	}

	log.Info("Upgrading system base")
	if err := pman.Upgrade(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to upgrade rootfs")
		return err
	}

	log.Info("Asserting system.devel component installation")
	if err := pman.InstallComponent("system.devel"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to assert system.devel")
		return err
	}

	// Ensure all directories are in place
	if err := p.CreateDirs(overlay); err != nil {
		return err
	}

	// Call the relevant build function
	if p.Type == PackageTypeYpkg {
		if err := p.BuildYpkg(notif, usr, pman, overlay, history); err != nil {
			return err
		}
	} else {
		if err := p.BuildXML(notif, pman, overlay); err != nil {
			return err
		}
	}

	return p.CollectAssets(overlay, usr)
}
