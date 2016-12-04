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
	_ "fmt"
	log "github.com/Sirupsen/logrus"
	_ "github.com/solus-project/libosdev/commands"
	_ "os"
)

// Index will attempt to index the given directory
func (p *Package) Index(notif PidNotifier, dir string, overlay *Overlay) error {
	log.WithFields(log.Fields{
		"profile": overlay.Back.Name,
	}).Info("Beginning indexer")

	ChrootEnvironment = SaneEnvironment("root", "/root")

	if err := p.ActivateRoot(overlay); err != nil {
		return err
	}

	// TODO: Bind mount and chroot!
	return nil
}
