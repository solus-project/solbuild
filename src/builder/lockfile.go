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
	"os"
	"sync"
	"syscall"
)

var (
	// ErrDeadLockFile is returned when an dead lockfile was encountered
	ErrDeadLockFile = errors.New("Dead lockfile")

	// ErrOwnedLockFile is returned when the lockfile is already owned by
	// another active process.
	ErrOwnedLockFile = errors.New("File is locked by someone else")
)

// A LockFile encapsulates locking functionality
type LockFile struct {
	path      string      // Path of the lockfile
	owningPID int         // Process ID of the lockfile owner
	ourPID    int         // Our own process ID
	conlock   *sync.Mutex // Concurrency lock for library use
}

// NewLockFile will return a new lockfile for the given path
func NewLockFile(path string) *LockFile {
	return &LockFile{
		path:      path,
		owningPID: -1,
		ourPID:    os.Getpid(),
		conlock:   new(sync.Mutex),
	}
}

// Lock will attempt to lock the file, or return an error if this fails
func (l *LockFile) Lock() error {
	l.conlock.Lock()
	defer l.conlock.Unlock()

	if PathExists(l.path) {
		if pid, err := l.readPID(); err == nil {
			// Unix this always works
			p, _ := os.FindProcess(pid)
			// Process is still active
			if err2 := p.Signal(syscall.Signal(0)); err2 != nil {
				if p.Pid != l.ourPID {
					l.owningPID = p.Pid
					return ErrOwnedLockFile
				}
			}
		} else if err != ErrDeadLockFile {
			return err
		}
	}
	return errors.New("Not yet implemented")
}

// Unlock will attempt to unlock the file, or return an error if this fails
func (l *LockFile) Unlock() error {
	l.conlock.Lock()
	defer l.conlock.Unlock()
	return errors.New("Not yet implemented")
}

// readPID is a simple utility to extract the PID from a file
func (l *LockFile) readPID() (int, error) {
	fi, err := os.Open(l.path)
	// Likely a permission issue.
	if err != nil {
		return -1, err
	}
	defer fi.Close()
	var pid int
	var n int

	// This is ok, we can just nuke it..
	if n, err = fmt.Fscanf(fi, "%d", &pid); err != nil {
		return -1, ErrDeadLockFile
	}
	// This is actually ok.
	if n != 1 {
		return -1, ErrDeadLockFile
	}
	return pid, nil
}
