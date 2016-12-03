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
	path      string        // Path of the lockfile
	owningPID int           // Process ID of the lockfile owner
	ourPID    int           // Our own process ID
	conlock   *sync.RWMutex // Concurrency lock for library use
	fd        *os.File      // Actual file being locked
}

// NewLockFile will return a new lockfile for the given path
func NewLockFile(path string) (*LockFile, error) {
	lock := &LockFile{
		path:      path,
		owningPID: -1,
		ourPID:    os.Getpid(),
		conlock:   new(sync.RWMutex),
		fd:        nil,
	}

	// We can consider setting the permissions to 0600
	w, err := os.OpenFile(lock.path, os.O_RDWR|os.O_CREATE, 00644)
	if err != nil {
		return nil, err
	}
	// Store the file descriptor
	lock.fd = w

	return lock, nil
}

// Lock will attempt to lock the file, or return an error if this fails
func (l *LockFile) Lock() error {
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
	l.conlock.Lock()

	// Finally lock it.
	if err := syscall.Flock(int(l.fd.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		l.conlock.Unlock()
		return err
	}

	l.conlock.Unlock()

	// Write the PID now we have an exclusive lock on it
	if err := l.writePID(); err != nil {
		return err
	}

	return nil
}

// Unlock will attempt to unlock the file, or return an error if this fails
func (l *LockFile) Unlock() error {
	if l.fd == nil {
		return errors.New("Cannot unlock that which we don't own!")
	}

	if err := syscall.Flock(int(l.fd.Fd()), syscall.LOCK_UN); err != nil {
		return err
	}
	return nil
}

// readPID is a simple utility to extract the PID from a file
func (l *LockFile) readPID() (int, error) {
	l.conlock.RLock()
	defer l.conlock.RUnlock()

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

// writePID will store our PID in the lockfile
func (l *LockFile) writePID() error {
	if l.fd == nil {
		panic(errors.New("Cannot write PID for no file!"))
	}
	l.conlock.Lock()
	defer l.conlock.Unlock()
	if _, err := fmt.Fprintf(l.fd, "%d", l.ourPID); err != nil {
		return err
	}
	return l.fd.Sync()
}

// Clean will dispose of the lock file and hopefully the lockfile itself
func (l *LockFile) Clean() error {
	l.conlock.Lock()
	defer l.conlock.Unlock()
	if l.fd == nil {
		return nil
	}

	// We don't own it
	if l.ourPID != l.owningPID {
		return nil
	}

	l.fd.Close()
	return os.Remove(l.path)
}
