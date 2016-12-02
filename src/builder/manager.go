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

	cancelled bool // Whether or not we've been cancelled

	activePID int // Active PID
}

// NewManager will return a newly initialised manager instance
func NewManager() *Manager {
	man := &Manager{
		cancelled: false,
		activePID: 0,
	}
	man.lock = new(sync.Mutex)
	return man
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

	if !IsValidProfile(profile) {
		return ErrInvalidProfile
	}

	if m.image != nil {
		return ErrManagerInitialised
	}

	m.image = NewBackingImage(profile)
	return nil
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
	m.overlay = NewOverlay(m.image, m.pkg)
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

	// Try to kill the active root PID first
	if m.activePID > 0 {
		syscall.Kill(m.activePID, syscall.SIGKILL)
		time.Sleep(2 * time.Second)
		syscall.Kill(m.activePID, syscall.SIGKILL)
		m.activePID = 0
	}

	// Still might have *something* alive in there, kill it with fire.
	for i := 0; i < 10; i++ {
		MurderDeathKill(m.overlay.MountPoint)
	}

	if m.pkg != nil {
		m.pkg.DeactivateRoot(m.overlay)
	}
	// Deactivation may have started something off, kill them too
	MurderDeathKill(m.overlay.MountPoint)

	// Unmount anything we may have mounted
	disk.GetMountManager().UnmountAll()
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

	return m.pkg.Build(m, m.pkgManager, m.overlay)
}
