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
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// PidNotifier provides a simple way to set the PID on a blocking process
type PidNotifier interface {
	SetActivePID(int)
}

// ActivateRoot will do the hard work of actually bring up the overlayfs
// system to allow manipulation of the roots for builds, etc.
func (p *Package) ActivateRoot(overlay *Overlay) error {
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
	overlay.Unmount()
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

// TouchFile will create the file if it doesn't exist, enabling use of bind
// mounts.
func TouchFile(path string) error {
	w, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 00644)
	if err != nil {
		return err
	}
	defer w.Close()
	return nil
}

// ChrootExec is a simple wrapper to return a correctly set up chroot command,
// so that we can store the PID, for long running tasks
func ChrootExec(notif PidNotifier, dir, command string) error {
	args := []string{dir, "/bin/sh", "-c", command}
	c := exec.Command("chroot", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = nil
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := c.Start(); err != nil {
		return err
	}
	notif.SetActivePID(c.Process.Pid)
	return c.Wait()
}

// ChrootExecStdin is almost identical to ChrootExec, except it permits a stdin
// to be associated with the command
func ChrootExecStdin(notif PidNotifier, dir, command string) error {
	args := []string{dir, "/bin/sh", "-c", command}
	c := exec.Command("chroot", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	if err := c.Start(); err != nil {
		return err
	}
	notif.SetActivePID(c.Process.Pid)
	return c.Wait()
}
