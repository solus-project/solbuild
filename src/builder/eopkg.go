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

// An EopkgRepo is a simplistic representation of a repo found in any given
// chroot.
type EopkgRepo struct {
	ID  string
	URI string
}

// EopkgManager is our own very shorted version of libosdev EopkgManager, to
// enable very very simple operations
type EopkgManager struct {
	dbusActive  bool
	root        string
	cacheSource string
	cacheTarget string
	dbusPid     string

	notif PidNotifier
}

// NewEopkgManager will return a new eopkg manager
func NewEopkgManager(notif PidNotifier, root string) *EopkgManager {
	return &EopkgManager{
		dbusActive:  false,
		root:        root,
		cacheSource: PackageCacheDirectory,
		cacheTarget: filepath.Join(root, "var/cache/eopkg/packages"),
		dbusPid:     filepath.Join(root, "var/run/dbus/pid"),
		notif:       notif,
	}
}

// CopyAssets will copy any required host-side assets into the system. This
// function has to be reusable simply because performing an eopkg upgrade
// or installing deps, prior to building, could clobber the files.
func (e *EopkgManager) CopyAssets() error {
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
	return nil
}

// Init will do some basic preparation of the chroot
func (e *EopkgManager) Init() error {
	// Ensure dbus pid is gone
	if PathExists(e.dbusPid) {
		if err := os.Remove(e.dbusPid); err != nil {
			return err
		}
	}

	if err := e.CopyAssets(); err != nil {
		return err
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
	dbusDir := filepath.Join(e.root, "run", "dbus")
	if err := os.MkdirAll(dbusDir, 00755); err != nil {
		return err
	}
	if err := ChrootExec(e.notif, e.root, "dbus-uuidgen --ensure"); err != nil {
		return err
	}
	e.notif.SetActivePID(0)
	if err := ChrootExec(e.notif, e.root, "dbus-daemon --system"); err != nil {
		return err
	}
	e.notif.SetActivePID(0)
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
	if err := ChrootExec(e.notif, e.root, "eopkg upgrade -y"); err != nil {
		return err
	}
	e.notif.SetActivePID(0)
	err := ChrootExec(e.notif, e.root, fmt.Sprintf("eopkg install -y %s", strings.Join(newReqs, " ")))
	return err
}

// InstallComponent will install the named component inside the chroot
func (e *EopkgManager) InstallComponent(comp string) error {
	err := ChrootExec(e.notif, e.root, fmt.Sprintf("eopkg install -c %v -y", comp))
	e.notif.SetActivePID(0)
	return err
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

	// Now we must nuke /run if it exists inside the chroot!
	runPath := filepath.Join(root, "run")
	if PathExists(runPath) {
		if err := os.RemoveAll(runPath); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to clean stale /run")
			return err
		}
	}

	if err := os.MkdirAll(runPath, 00755); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to clean create /run")
		return err
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

// Read the given plaintext URI file to find the target
func readURIFile(path string) (string, error) {
	fi, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer fi.Close()
	contents, err := ioutil.ReadAll(fi)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

// GetRepos will attempt to discover all the repos on the target filesystem
func (e *EopkgManager) GetRepos() ([]*EopkgRepo, error) {
	globPat := filepath.Join(e.root, "var", "lib", "eopkg", "index", "*", "uri")
	var repoFiles []string

	repoFiles, _ = filepath.Glob(globPat)
	// No repos
	if len(repoFiles) < 1 {
		return nil, nil
	}

	var repos []*EopkgRepo

	for _, repo := range repoFiles {
		uri, err := readURIFile(repo)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"path":  repo,
			}).Error("Unable to read repository file")
			return nil, err
		}
		repoName := filepath.Base(filepath.Dir(repo))
		repos = append(repos, &EopkgRepo{
			ID:  repoName,
			URI: uri,
		})
	}
	return repos, nil
}

// AddRepo will attempt to add a repo to the filesystem
func (e *EopkgManager) AddRepo(id, source string) error {
	e.notif.SetActivePID(0)
	return ChrootExec(e.notif, e.root, fmt.Sprintf("eopkg add-repo '%s' '%s'", id, source))
}

// RemoveRepo will attempt to remove a named repo from the filesystem
func (e *EopkgManager) RemoveRepo(id string) error {
	e.notif.SetActivePID(0)
	return ChrootExec(e.notif, e.root, fmt.Sprintf("eopkg remove-repo '%s'", id))
}
