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
	"github.com/solus-project/libosdev/commands"
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
}

// NewEopkgManager will return a new eopkg manager
func NewEopkgManager(root string) *EopkgManager {
	return &EopkgManager{
		dbusActive:  false,
		root:        root,
		cacheSource: PackageCacheDirectory,
		cacheTarget: filepath.Join(root, "var/cache/eopkg/packages"),
	}
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
	fpath := filepath.Join(e.root, "var/run/dbus/pid")
	var b []byte
	var err error
	var f *os.File

	if f, err = os.Open(fpath); err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(fpath)
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
}

// Upgrade will perform an eopkg upgrade inside the chroot
func (e *EopkgManager) Upgrade() error {
	return commands.ChrootExec(e.root, "eopkg upgrade -y")
}

// InstallComponent will install the named component inside the chroot
func (e *EopkgManager) InstallComponent(comp string) error {
	return commands.ChrootExec(e.root, fmt.Sprintf("eopkg install -c %v -y", comp))
}
