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

var (
	// ErrCannotContinue is a stock error return
	ErrCannotContinue = errors.New("Index cannot continue")

	// IndexBindTarget is where we always mount the repo
	IndexBindTarget = "/hostRepo/Index"
)

// Index will attempt to index the given directory
func (p *Package) Index(notif PidNotifier, dir string, overlay *Overlay) error {
	log.WithFields(log.Fields{
		"profile": overlay.Back.Name,
	}).Info("Beginning indexer")

	mman := disk.GetMountManager()

	ChrootEnvironment = SaneEnvironment("root", "/root")

	// Check the source exists first!
	if !PathExists(dir) {
		log.WithFields(log.Fields{
			"dir": dir,
		}).Error("Directory does not exist")
		return ErrCannotContinue
	}

	// Indexer will always create new dirs..
	if err := overlay.CleanExisting(); err != nil {
		return err
	}

	if err := p.ActivateRoot(overlay); err != nil {
		return err
	}

	// Create the target
	target := filepath.Join(overlay.MountPoint, IndexBindTarget[:1])
	if err := os.MkdirAll(target, 00755); err != nil {
		log.WithFields(log.Fields{
			"dir":   target,
			"error": err,
		}).Error("Cannot create bind target")
		return err
	}

	log.WithFields(log.Fields{
		"dir": dir,
	}).Info("Bind mounting directory for indexing")

	if err := mman.BindMount(dir, target); err != nil {
		log.WithFields(log.Fields{
			"dir":   target,
			"error": err,
		}).Error("Cannot bind mount directory")
		return err
	}

	// Ensure it gets cleaned up
	overlay.ExtraMounts = append(overlay.ExtraMounts, target)

	log.Info("Now indexing")
	command := fmt.Sprintf("cd %s; eopkg index --skip-signing .", IndexBindTarget)
	if err := ChrootExec(notif, overlay.MountPoint, command); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"dir":   dir,
		}).Error("Indexing failed")
		return err
	}
	return nil
}
