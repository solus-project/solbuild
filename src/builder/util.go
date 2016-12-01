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
	log "github.com/Sirupsen/logrus"
	"github.com/solus-project/libosdev/commands"
	"github.com/solus-project/libosdev/disk"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// ActivateRoot will do the hard work of actually bring up the overlayfs
// system to allow manipulation of the roots for builds, etc.
func (p *Package) ActivateRoot(overlay *Overlay) error {
	// First things first, setup the namespace
	if err := ConfigureNamespace(); err != nil {
		return err
	}

	log.Info("Configuring overlay storage")

	// Set up environment
	if err := overlay.EnsureDirs(); err != nil {
		return err
	}

	// Now mount the overlayfs
	if err := overlay.Mount(); err != nil {
		return err
	}

	// Add build user
	if p.Type == PackageTypeYpkg {
		if err := overlay.AddBuildUser(); err != nil {
			return err
		}
	}

	log.Info("Bringing up virtual filesystems")
	if err := overlay.MountVFS(); err != nil {
		return err
	}
	return nil
}

// DeactivateRoot will tear down the previously activated root
func (p *Package) DeactivateRoot(overlay *Overlay) {
	MurderDeathKill(overlay.MountPoint)
	mountMan := disk.GetMountManager()
	commands.SetStdin(nil)
	if err := overlay.Unmount(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error unmounting overlay")
	}
	log.Info("Requesting unmount of all remaining mountpoints")
	mountMan.UnmountAll()
}

// MurderDeathKill will find all processes with a root matching the given root
// and set about killing them, to assist in clean closing.
func MurderDeathKill(root string) error {
	path, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}

	var files []os.FileInfo

	if files, err = ioutil.ReadDir("/proc"); err != nil {
		return err
	}

	for _, f := range files {
		fpath := filepath.Join("/proc", f.Name(), "cwd")

		spath, err := filepath.EvalSymlinks(fpath)
		if err != nil {
			continue
		}

		if spath != path {
			continue
		}

		spid := f.Name()
		var pid int

		if pid, err = strconv.Atoi(spid); err != nil {
			log.WithFields(log.Fields{
				"pid":   spid,
				"error": err,
			}).Error("POSIX Weeps - broken pid identifier")
			return err
		}

		log.WithFields(log.Fields{
			"pid": pid,
		}).Info("Killing child process in chroot")

		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			log.WithFields(log.Fields{
				"pid": pid,
			}).Error("Error terminating process, attempting force kill")
			time.Sleep(400 * time.Millisecond)
			if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
				log.WithFields(log.Fields{
					"pid": pid,
				}).Error("Error killing (-9) process")
			}
		}
	}
	return nil
}

// HandleInterrupt will call the specified reaper function when the terminal
// is interrupted (i.e. ctrl+c)
func HandleInterrupt(v func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		log.Warning("CTRL+C interrupted, cleaning up")
		v()
		os.Exit(1)
	}()
}
