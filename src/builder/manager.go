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
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	// ErrManagerInitialised is returned when the library user attempts to set
	// a core part of the Manager after it's already been initialised
	ErrManagerInitialised = errors.New("The manager has already been initialised")

	// ErrNoPackage is returned when we've got no package
	ErrNoPackage = errors.New("You must first set a package to build it")

	// ErrNotImplemented is returned as a placeholder when developing functionality.
	ErrNotImplemented = errors.New("Function not yet implemented")

	// ErrProfileNotInstalled is returned when a profile is not yet installed
	ErrProfileNotInstalled = errors.New("Profile is not installed")

	// ErrInvalidProfile is returned when there is an invalid profile
	ErrInvalidProfile = errors.New("Invalid profile")

	// ErrInvalidImage is returned when the backing image is unknown
	ErrInvalidImage = errors.New("Invalid image")

	// ErrInterrupted is returned when the build is interrupted
	ErrInterrupted = errors.New("The operation was cancelled by the user")
)

// A Manager is responsible for cleanly managing the entire session within solbuild,
// i.e. setup, teardown, cleaning up, etc.
//
// The consumer should create a new manager instance and only use these methods,
// not bypass and use API methods.
type Manager struct {
	image      *BackingImage // Storage for the overlay
	overlay    *Overlay      // OverlayFS configuration
	pkg        *Package      // Current package, if any
	pkgManager *EopkgManager // Package manager, if any
	lock       *sync.Mutex   // Lock on all operations to prevent.. damage.
	profile    *Profile      // The profile we've been requested to use

	lockfile *LockFile // We track the global lock for each operation
	didStart bool      // Whether we got anything done.

	cancelled  bool // Whether or not we've been cancelled
	updateMode bool // Whether we're just updating an image

	config *Config // Our config from the merged system/vendor configs

	activePID int // Active PID
}

// NewManager will return a newly initialised manager instance
func NewManager() (*Manager, error) {
	// First things first, setup the namespace
	if err := ConfigureNamespace(); err != nil {
		return nil, err
	}
	man := &Manager{
		cancelled:  false,
		activePID:  0,
		updateMode: false,
		lockfile:   nil,
		didStart:   false,
	}

	// Now load the configuration in
	if config, err := NewConfig(); err == nil {
		man.config = config
	} else {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to load solbuild configuration")
		return nil, err
	}

	man.lock = new(sync.Mutex)
	return man, nil
}

// SetActivePID will set the active task PID
func (m *Manager) SetActivePID(pid int) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.activePID = pid
}

// SetProfile will attempt to initialise the manager with a given profile
// Currently this is locked to a backing image specification, but in future
// will be expanded to support profiles *based* on backing images.
func (m *Manager) SetProfile(profile string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Passed an empty profile from the CLI flags, so set our default profile
	// as the one to use.
	if profile == "" {
		profile = m.config.DefaultProfile
	}

	prof, err := NewProfile(profile)
	if err != nil {
		EmitProfileError(profile)
		return err
	}

	if !IsValidImage(prof.Image) {
		EmitImageError(prof.Image)
		return ErrInvalidImage
	}

	if m.image != nil {
		return ErrManagerInitialised
	}

	m.profile = prof
	m.image = NewBackingImage(m.profile.Image)
	return nil
}

// GetProfile will return the profile associated with this builder
func (m *Manager) GetProfile() *Profile {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.profile
}

// SetPackage will set the package associated with this manager.
// This package will be used in build & chroot operations only.
func (m *Manager) SetPackage(pkg *Package) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.pkg != nil {
		return ErrManagerInitialised
	}

	if !m.image.IsInstalled() {
		return ErrProfileNotInstalled
	}

	m.pkg = pkg
	m.overlay = NewOverlay(m.profile, m.image, m.pkg)
	m.pkgManager = NewEopkgManager(m, m.overlay.MountPoint)
	return nil
}

// IsCancelled will determine if the build has been cancelled, this will result
// in a lot of locking between all operations
func (m *Manager) IsCancelled() bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.cancelled
}

// SetCancelled will mark the build manager as cancelled, so it should not attempt
// to start any new operations whatsoever.
func (m *Manager) SetCancelled() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.cancelled = true
}

// Cleanup will take care of any teardown operations. It takes an exclusive lock
// and ensures all cleaning is handled before anyone else is permitted to continue,
// at which point error propagation and the IsCancelled() function should be enough
// logic to go on.
func (m *Manager) Cleanup() {
	if !m.didStart {
		return
	}
	log.Debug("Acquiring global lock")
	m.lock.Lock()
	defer m.lock.Unlock()
	log.Info("Cleaning up")

	if m.pkgManager != nil {
		// Potentially unnecessary but meh
		m.pkgManager.StopDBUS()
		// Always needed
		m.pkgManager.Cleanup()
	}

	deathPoint := ""
	if m.overlay != nil {
		deathPoint = m.overlay.MountPoint
	}
	if m.updateMode {
		deathPoint = m.image.RootDir
	}

	// Try to kill the active root PID first
	if m.activePID > 0 {
		syscall.Kill(-m.activePID, syscall.SIGKILL)
		time.Sleep(2 * time.Second)
		syscall.Kill(-m.activePID, syscall.SIGKILL)
		m.activePID = 0
	}

	// Still might have *something* alive in there, kill it with fire.
	if deathPoint != "" {
		for i := 0; i < 10; i++ {
			MurderDeathKill(deathPoint)
		}
	}

	if m.pkg != nil {
		m.pkg.DeactivateRoot(m.overlay)
	}

	// Deactivation may have started something off, kill them too
	if deathPoint != "" {
		MurderDeathKill(deathPoint)
	}

	// Unmount anything we may have mounted
	disk.GetMountManager().UnmountAll()

	// Finally clean out the lock files
	if m.lockfile != nil {
		if err := m.lockfile.Unlock(); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failure in unlocking root")
		}
		if err := m.lockfile.Clean(); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failure in cleaning lockfile")
		}
	}
}

// doLock will handle the relevant locking operation for the given path
func (m *Manager) doLock(path, opType string) error {
	// Handle file locking
	lock, err := NewLockFile(path)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"file":  path,
		}).Error("Failed to lock root for " + opType)
		return err
	}
	m.lockfile = lock

	if err = m.lockfile.Lock(); err != nil {
		if err == ErrOwnedLockFile {
			log.WithFields(log.Fields{
				"error":   err,
				"pid":     m.lockfile.GetOwnerPID(),
				"process": m.lockfile.GetOwnerProcess(),
			}).Error("Failed to lock root - another process is using it")
		} else {
			log.WithFields(log.Fields{
				"error": err,
				"pid":   m.lockfile.GetOwnerPID(),
			}).Error("Failed to lock root")
		}
		return err
	}
	m.didStart = true
	return nil
}

// SigIntCleanup will take care of cleaning up the build process.
func (m *Manager) SigIntCleanup() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		log.Warning("CTRL+C interrupted, cleaning up")
		m.SetCancelled()
		m.Cleanup()
		log.Error("Exiting due to interruption")
		os.Exit(1)
	}()
}

// Build will attempt to build the package associated with this manager,
// automatically handling any required cleanups.
func (m *Manager) Build() error {
	if m.IsCancelled() {
		return ErrInterrupted
	}

	m.lock.Lock()
	if m.pkg == nil {
		m.lock.Unlock()
		return ErrNoPackage
	}
	m.lock.Unlock()

	// Now get on with the real work!
	defer m.Cleanup()
	m.SigIntCleanup()

	// Now set our options according to the config
	m.overlay.EnableTmpfs = m.config.EnableTmpfs
	m.overlay.TmpfsSize = m.config.TmpfsSize

	if err := m.doLock(m.overlay.LockPath, "building"); err != nil {
		return err
	}

	return m.pkg.Build(m, m.GetProfile(), m.pkgManager, m.overlay)
}

// Chroot will enter the build environment to allow users to introspect it
func (m *Manager) Chroot() error {
	if m.IsCancelled() {
		return ErrInterrupted
	}

	m.lock.Lock()
	if m.pkg == nil {
		m.lock.Unlock()
		return ErrNoPackage
	}
	m.lock.Unlock()

	// Now get on with the real work!
	defer m.Cleanup()
	m.SigIntCleanup()

	if err := m.doLock(m.overlay.LockPath, "chroot"); err != nil {
		return err
	}

	return m.pkg.Chroot(m, m.pkgManager, m.overlay)
}

// Update will attempt to update the base image
func (m *Manager) Update() error {
	if m.IsCancelled() {
		return ErrInterrupted
	}
	m.lock.Lock()
	if m.image == nil {
		m.lock.Unlock()
		return ErrInvalidProfile
	}
	if !m.image.IsInstalled() {
		m.lock.Unlock()
		return ErrProfileNotInstalled
	}
	m.updateMode = true
	m.pkgManager = NewEopkgManager(m, m.image.RootDir)
	m.lock.Unlock()

	defer m.Cleanup()
	m.SigIntCleanup()

	if err := m.doLock(m.image.LockPath, "updating"); err != nil {
		return err
	}

	return m.image.Update(m, m.pkgManager)
}

// SetTmpfs sets the manager tmpfs option
func (m *Manager) SetTmpfs(enable bool, size string) {
	if m.IsCancelled() {
		return
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.overlay != nil {
		m.overlay.EnableTmpfs = enable
		m.overlay.TmpfsSize = strings.TrimSpace(size)
	}
}
