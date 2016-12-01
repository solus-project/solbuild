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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// EopkgManager is our own very shorted version of libosdev EopkgManager, to
// enable very very simple operations
type EopkgManager struct {
	dbusActive  bool
	root        string
	cacheSource string
	cacheTarget string
	dbusPid     string
}

// NewEopkgManager will return a new eopkg manager
func NewEopkgManager(root string) *EopkgManager {
	return &EopkgManager{
		dbusActive:  false,
		root:        root,
		cacheSource: PackageCacheDirectory,
		cacheTarget: filepath.Join(root, "var/cache/eopkg/packages"),
		dbusPid:     filepath.Join(root, "var/run/dbus/pid"),
	}
}

// Init will do some basic preparation of the chroot
func (e *EopkgManager) Init() error {
	// Ensure dbus pid is gone
	if PathExists(e.dbusPid) {
		if err := os.Remove(e.dbusPid); err != nil {
			return err
		}
	}

	requiredAssets := map[string]string{
		"/etc/resolv.conf":      filepath.Join(e.root, "etc/resolv.conf"),
		"/etc/eopkg/eopkg.conf": filepath.Join(e.root, "etc/eopkg/eopkg.conf"),
	}

	for key, value := range requiredAssets {
		if !PathExists(key) {
			continue
		}
		dirName := filepath.Dir(value)
		if !PathExists(dirName) {
			log.WithFields(log.Fields{
				"dir": dirName,
			}).Debug("Creating required directory")
			if err := os.MkdirAll(dirName, 00755); err != nil {
				log.WithFields(log.Fields{
					"dir":   dirName,
					"error": err,
				}).Error("Failed to create required asset directory")
				return err
			}
		}
		log.WithFields(log.Fields{
			"file": key,
		}).Debug("Copying host asset")
		if err := disk.CopyFile(key, value); err != nil {
			log.WithFields(log.Fields{
				"file":  key,
				"error": err,
			}).Error("Failed to copy host asset")
			return err
		}
	}

	// Ensure system wide cache exists
	if !PathExists(e.cacheSource) {
		log.WithFields(log.Fields{
			"dir": e.cacheSource,
		}).Debug("Creating system-wide package cache")
		if err := os.MkdirAll(e.cacheSource, 00755); err != nil {
			log.WithFields(log.Fields{
				"dir":   e.cacheSource,
				"error": err,
			}).Error("Failed to create package cache")
			return err
		}
	}

	if err := os.MkdirAll(e.cacheTarget, 00755); err != nil {
		return err
	}
	return disk.GetMountManager().BindMount(e.cacheSource, e.cacheTarget)
}

// StartDBUS will bring up dbus within the chroot
func (e *EopkgManager) StartDBUS() error {
	if e.dbusActive {
		return nil
	}
	if err := commands.ChrootExec(e.root, "dbus-uuidgen --ensure"); err != nil {
		return err
	}
	if err := commands.ChrootExec(e.root, "dbus-daemon --system"); err != nil {
		return err
	}
	e.dbusActive = true
	return nil
}

// StopDBUS will tear down dbus
func (e *EopkgManager) StopDBUS() error {
	// No sense killing dbus twice
	if !e.dbusActive {
		return nil
	}
	var b []byte
	var err error
	var f *os.File

	if f, err = os.Open(e.dbusPid); err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(e.dbusPid)
		e.dbusActive = false
	}()

	if b, err = ioutil.ReadAll(f); err != nil {
		return err
	}

	pid := strings.Split(string(b), "\n")[0]
	return commands.ExecStdoutArgs("kill", []string{"-9", pid})
}

// Cleanup will take care of any work we've already done before
func (e *EopkgManager) Cleanup() {
	e.StopDBUS()
	disk.GetMountManager().Unmount(e.cacheTarget)
}

// Upgrade will perform an eopkg upgrade inside the chroot
func (e *EopkgManager) Upgrade() error {
	// Certain requirements may not be in system.base, but are required for
	// proper containerized functionality.
	newReqs := []string{
		"iproute2",
	}
	if err := commands.ChrootExec(e.root, "eopkg upgrade -y"); err != nil {
		return err
	}
	return commands.ChrootExec(e.root, fmt.Sprintf("eopkg install -y %s", strings.Join(newReqs, " ")))
}

// InstallComponent will install the named component inside the chroot
func (e *EopkgManager) InstallComponent(comp string) error {
	return commands.ChrootExec(e.root, fmt.Sprintf("eopkg install -c %v -y", comp))
}

// EnsureEopkgLayout will enforce changes to the filesystem to make sure that
// it works as expected.
func EnsureEopkgLayout(root string) error {
	// Ensures we don't end up with /var/lock vs /run/lock nonsense
	reqDirs := []string{
		"run/lock",
		"var",
		// Enables our bind mounting for caching
		"var/cache/eopkg/packages",
	}

	// Construct the required directories in the tree
	for _, dir := range reqDirs {
		dirPath := filepath.Join(root, dir)
		if err := os.MkdirAll(dirPath, 00755); err != nil {
			return err
		}
	}

	lockTgt := filepath.Join(root, "var", "lock")
	if !PathExists(lockTgt) {
		if err := os.Symlink("../run/lock", lockTgt); err != nil {
			return err
		}
	}
	runTgt := filepath.Join(root, "var", "run")
	if !PathExists(runTgt) {
		if err := os.Symlink("../run", filepath.Join(root, "var", "run")); err != nil {
			return err
		}
	}

	return nil
}
