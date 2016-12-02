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
	"sync"
)

var (
	// ErrManagerInitialised is returned when the library user attempts to set
	// a core part of the Manager after it's already been initialised
	ErrManagerInitialised = errors.New("The manager has already been initialised")

	// ErrNoPackage is returned when we've got no package
	ErrNoPackage = errors.New("You must first set a package to build it")

	// ErrNotImplemented is returned as a placeholder when developing functionality.
	ErrNotImplemented = errors.New("Function not yet implemented")
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
}

// NewManager will return a newly initialised manager instance
func NewManager() *Manager {
	man := &Manager{}
	man.lock = new(sync.Mutex)
	return man
}

// SetProfile will attempt to initialise the manager with a given profile
// Currently this is locked to a backing image specification, but in future
// will be expanded to support profiles *based* on backing images.
func (m *Manager) SetProfile(profile string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if !IsValidProfile(profile) {
		return fmt.Errorf("Invalid profile: %v", profile)
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
	m.pkg = pkg
	return nil
}

// Build will attempt to build the package associated with this manager,
// automatically handling any required cleanups.
func (m *Manager) Build() error {
	return ErrNotImplemented
}
