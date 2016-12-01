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
	log "github.com/Sirupsen/logrus"
	"github.com/solus-project/libosdev/commands"
	"github.com/solus-project/libosdev/disk"
)

// ActivateRoot will do the hard work of actually bring up the overlayfs
// system to allow manipulation of the roots for builds, etc.
func (p *Package) ActivateRoot(overlay *Overlay) error {
	// First things first, setup the namespace
	if err := ConfigureNamespace(); err != nil {
		return err
	}

	log.Info("Configuring overlay storage")

	// Set up environment
	if err := overlay.EnsureDirs(); err != nil {
		return err
	}

	// Now mount the overlayfs
	if err := overlay.Mount(); err != nil {
		return err
	}

	// Add build user
	if p.Type == PackageTypeXML {
		if err := overlay.AddBuildUser(); err != nil {
			return err
		}
	}

	log.Info("Bringing up virtual filesystems")
	if err := overlay.MountVFS(); err != nil {
		return err
	}
	return nil
}

// DeactivateRoot will tear down the previously activated root
func (p *Package) DeactivateRoot(overlay *Overlay) {
	mountMan := disk.GetMountManager()
	commands.SetStdin(nil)
	if err := overlay.Unmount(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error unmounting overlay")
	}
	log.Info("Requesting unmount of all remaining mountpoints")
	mountMan.UnmountAll()
}
