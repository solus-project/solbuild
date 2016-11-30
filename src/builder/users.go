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

// A User is an /etc/passwd defined user
type User struct {
	Name  string
	UID   int
	GID   int
	Gecos string
	Home  string
	Shell string
}

// A Group is an /etc/group defined user
type Group struct {
	GID int
}

// Passwd is a simple helper to parse passwd files from a chroot
type Passwd struct {
	Users  map[string]*User
	Groups map[int]*Group
}

// NewPasswd will parse the given path and return a friendly representation
// of those files
func NewPasswd(path string) (*Passwd, error) {
	return nil, errors.New("Not yet implemented")
}
