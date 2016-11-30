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
	"os"
)

// Update will attempt to update the backing image to the latest version
// internally.
func (b *BackingImage) Update() error {
	// TODO: Check if it is locked!

	mountMan := disk.GetMountManager()
	defer mountMan.UnmountAll()

	log.WithFields(log.Fields{
		"image": b.Name,
	}).Info("Updating backing image")

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

	return errors.New("Not yet implemented")
}
