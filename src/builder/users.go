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
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// A User is an /etc/passwd defined user
type User struct {
	Name  string // User Name
	UID   int    // User ID
	GID   int    // User primary Group ID
	Gecos string // User Gecos (Pretty name)
	Home  string // User home directory
	Shell string // User shell program
}

// A Group is an /etc/group defined user
type Group struct {
	Name    string   // Group Name
	ID      int      // Group ID
	Members []string // Names of users in group
}

// Passwd is a simple helper to parse passwd files from a chroot
type Passwd struct {
	Users  map[string]*User
	Groups map[string]*Group
}

// NewPasswd will parse the given path and return a friendly representation
// of those files
func NewPasswd(path string) (*Passwd, error) {
	passwdPath := filepath.Join(path, "passwd")
	groupPath := filepath.Join(path, "group")

	var err error

	ret := &Passwd{}
	if ret.Users, err = ParseUsers(passwdPath); err != nil {
		return nil, err
	}
	if ret.Groups, err = ParseGroups(groupPath); err != nil {
		return nil, err
	}
	return ret, nil
}

// ParseUsers will attempt to parse a *NIX style passwd file
func ParseUsers(passwd string) (map[string]*User, error) {
	fi, err := os.Open(passwd)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	ret := make(map[string]*User)

	sc := bufio.NewScanner(fi)
	for sc.Scan() {
		line := sc.Text()
		splits := strings.Split(line, ":")
		if len(splits) != 7 {
			return nil, fmt.Errorf("Invalid number of fields in passwd file: %d", len(splits))
		}
		user := &User{
			Name:  strings.TrimSpace(splits[0]),
			Gecos: strings.TrimSpace(splits[4]),
			Home:  strings.TrimSpace(splits[5]),
			Shell: strings.TrimSpace(splits[6]),
		}
		// Parse the uid/gid
		if uid, err := strconv.Atoi(strings.TrimSpace(splits[2])); err == nil {
			user.UID = uid
		} else {
			return nil, err
		}
		if gid, err := strconv.Atoi(strings.TrimSpace(splits[3])); err == nil {
			user.GID = gid
		} else {
			return nil, err
		}
		// Success
		ret[user.Name] = user
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

// ParseGroups will attempt to parse a *NIX style group file
func ParseGroups(grps string) (map[string]*Group, error) {
	fi, err := os.Open(grps)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	ret := make(map[string]*Group)

	sc := bufio.NewScanner(fi)
	for sc.Scan() {
		line := sc.Text()
		splits := strings.Split(line, ":")
		if len(splits) != 4 {
			return nil, fmt.Errorf("Invalid number of fields in group file: %d", len(splits))
		}
		group := &Group{
			Name: strings.TrimSpace(splits[0]),
		}
		// So we don't get one empty member situation
		membs := strings.TrimSpace(splits[3])
		if membs != "" {
			group.Members = strings.Split(membs, ",")
		}
		// Parse the gid
		if gid, err := strconv.Atoi(strings.TrimSpace(splits[2])); err == nil {
			group.ID = gid
		} else {
			return nil, err
		}
		// Success
		ret[group.Name] = group
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}
