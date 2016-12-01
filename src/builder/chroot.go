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
	"os"
)

// Chroot will attempt to spawn a chroot in the overlayfs system
func (p *Package) Chroot(img *BackingImage) error {
	log.WithFields(log.Fields{
		"profile": img.Name,
		"version": p.Version,
		"package": p.Name,
		"type":    p.Type,
		"release": p.Release,
	}).Info("Beginning chroot")

	overlay := NewOverlay(img, p)

	defer p.DeactivateRoot(overlay)
	if err := p.ActivateRoot(overlay); err != nil {
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

	// TODO: Stay as root for pspec
	log.Info("Spawning login shell")
	// Allow bash to work
	commands.SetStdin(os.Stdin)
	loginCommand := fmt.Sprintf("/bin/su - %s -s %s", BuildUser, BuildUserShell)
	err := commands.ChrootExec(overlay.MountPoint, loginCommand)
	commands.SetStdin(nil)
	return err
}
