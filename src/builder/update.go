//
// Copyright © 2016-2017 Ikey Doherty <ikey@solus-project.com>
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
	log "github.com/Sirupsen/logrus"
	"github.com/solus-project/libosdev/disk"
	"os"
)

func (b *BackingImage) updatePackages(notif PidNotifier, pkgManager *EopkgManager) error {
	log.Debug("Initialising package manager")

	if err := pkgManager.Init(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to initialise package manager")
		return err
	}

	// Bring up dbus to do Things
	log.Debug("Starting D-BUS")
	if err := pkgManager.StartDBUS(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to start d-bus")
		return err
	}

	log.Debug("Upgrading builder image")
	if err := pkgManager.Upgrade(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to perform upgrade")
		return err
	}

	log.Debug("Asserting system.devel component")
	if err := pkgManager.InstallComponent("system.devel"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to install system.devel")
		return err
	}

	// Cleanup now
	log.Debug("Stopping D-BUS")
	if err := pkgManager.StopDBUS(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to stop d-bus")
		return err
	}

	return nil
}

// Update will attempt to update the backing image to the latest version
// internally.
func (b *BackingImage) Update(notif PidNotifier, pkgManager *EopkgManager) error {
	mountMan := disk.GetMountManager()
	log.WithFields(log.Fields{
		"image": b.Name,
	}).Debug("Updating backing image")

	if !PathExists(b.RootDir) {
		if err := os.MkdirAll(b.RootDir, 00755); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to create required directories")
			return err
		}
		log.WithFields(log.Fields{
			"dir": b.RootDir,
		}).Debug("Created root directory")
	}

	log.WithFields(log.Fields{
		"image": b.ImagePath,
		"root":  b.RootDir,
	}).Debug("Mounting rootfs")

	// Mount the rootfs
	if err := mountMan.Mount(b.ImagePath, b.RootDir, "auto", "loop"); err != nil {
		log.WithFields(log.Fields{
			"image": b.ImagePath,
			"error": err,
		}).Error("Failed to mount rootfs")
		return err
	}

	if err := EnsureEopkgLayout(b.RootDir); err != nil {
		log.WithFields(log.Fields{
			"image": b.ImagePath,
			"error": err,
		}).Error("Failed to fix filesystem layout")
		return err
	}

	// Hand over to package management to do the updates
	if err := b.updatePackages(notif, pkgManager); err != nil {
		return err
	}

	// Lastly, add the user
	if err := AddBuildUser(b.RootDir); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"profile": b.Name,
	}).Debug("Image successfully updated")

	return nil
}
