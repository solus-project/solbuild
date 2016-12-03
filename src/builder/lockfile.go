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
)

// A LockFile encapsulates locking functionality
type LockFile struct {
	path string
}

// NewLockFile will return a new lockfile for the given path
func NewLockFile(path string) *LockFile {
	return &LockFile{path: path}
}

// Lock will attempt to lock the file, or return an error if this fails
func (l *LockFile) Lock() error {
	return errors.New("Not yet implemented")
}

// Unlock will attempt to unlock the file, or return an error if this fails
func (l *LockFile) Unlock() error {
	return errors.New("Not yet implemented")
}
