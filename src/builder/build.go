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
	log "github.com/Sirupsen/logrus"
	"path/filepath"
)

// CopyAssets will copy all of the required assets into the builder root
func (p *Package) CopyAssets(o *Overlay) error {
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
	destdir := filepath.Join(o.MountPoint, "WORK")

	for _, p := range copyPaths {
		fso := filepath.Join(baseDir, p)
		if err := CopyAll(fso, destdir); err != nil {
			return err
		}
	}
	return nil
}

// Build will attempt to build the package in the overlayfs system
func (p *Package) Build(img *BackingImage) error {
	log.WithFields(log.Fields{
		"profile": img.Name,
		"version": p.Version,
		"package": p.Name,
		"type":    p.Type,
		"release": p.Release,
	}).Info("Building package")

	overlay := NewOverlay(img, p)

	// Set up environment
	if err := overlay.CleanExisting(); err != nil {
		return err
	}

	pman := NewEopkgManager(overlay.MountPoint)

	// Ensure we clean up after ourselves
	reaper := GrimReaper(overlay, p, pman)
	defer reaper()
	HandleInterrupt(reaper)

	// Bring up the root
	if err := p.ActivateRoot(overlay); err != nil {
		return err
	}

	// Ensure source assets are in place
	if err := p.CopyAssets(overlay); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to copy required source assets")
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

	if p.Type == PackageTypeYpkg {
		// TODO: Install build dependencies here.

		// Cleanup now
		log.Debug("Stopping D-BUS")
		if err := pman.StopDBUS(); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to stop d-bus")
			return err
		}

		// Now kill networking
		if err := DropNetworking(); err != nil {
			return nil
		}

		// Ensure the overlay can network on localhost only
		if err := overlay.ConfigureNetworking(); err != nil {
			return nil
		}
		// Now build the package
	} else {
		// Just straight up build it with eopkg
		log.Warning("Full sandboxing is not possible with legacy format")

		// Now we can stop dbus..
		log.Debug("Stopping D-BUS")
		if err := pman.StopDBUS(); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to stop d-bus")
			return err
		}
	}

	// TODO: Collect build results

	return errors.New("Not yet implemented")
}
