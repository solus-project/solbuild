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
	"github.com/solus-project/libosdev/disk"
)

// Build will attempt to build the package in the overlayfs system
func (p *Package) Build(img *BackingImage) error {
	// First things first, setup the namespace
	if err := ConfigureNamespace(); err != nil {
		return err
	}

	mountMan := disk.GetMountManager()
	overlay := NewOverlay(img, p)

	defer func() {
		if err := overlay.Unmount(); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Error unmounting overlay")
		}
		log.Info("Requesting unmount of all remaining mountpoints")
		mountMan.UnmountAll()
	}()

	log.WithFields(log.Fields{
		"profile": img.Name,
		"version": p.Version,
		"package": p.Name,
		"type":    p.Type,
		"release": p.Release,
	}).Info("Building package")

	// Warn about lack of sandboxing
	if p.Type != PackageTypeYpkg {
		log.Warning("Full sandboxing is not possible with legacy format")
	}

	log.Info("Configuring overlay storage")

	// Set up environment
	if err := overlay.CleanExisting(); err != nil {
		return err
	}
	if err := overlay.EnsureDirs(); err != nil {
		return err
	}

	// Now mount the overlayfs
	if err := overlay.Mount(); err != nil {
		return err
	}

	return errors.New("Not yet implemented")
}
