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
	"github.com/go-ini/ini"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

// UserInfo is required for ypkg builds, to set the .solus/package internally
// and propagate the author details.
type UserInfo struct {
	Name     string // Actual name
	Email    string // Actual email
	UID      int    // Unix User Id
	GID      int    // Unix Group ID
	HomeDir  string // Home directory of the user
	Username string // Textual username
}

const (
	// FallbackUserName is what we fallback to if everything else fails
	FallbackUserName = "Automated Package Build"

	// FallbackUserEmail is what we fallback to if everything else fails
	FallbackUserEmail = "no.email.set.in.config"
)

// SetFromSudo will attempt to set our details from sudo user environment
func (u *UserInfo) SetFromSudo() bool {
	sudoUID := os.Getenv("SUDO_UID")
	sudoGID := os.Getenv("SUDO_GID")
	uid := -1
	gid := -1
	var err error

	if sudoGID == "" {
		sudoGID = sudoUID
	}

	if sudoUID == "" {
		return false
	}

	if uid, err = strconv.Atoi(sudoUID); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"uid":   sudoUID,
		}).Error("Malformed SUDO_UID in environment")
		return false
	}

	if gid, err = strconv.Atoi(sudoGID); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"gid":   sudoGID,
		}).Error("Malformed SUDO_GID in environment")
		return false
	}

	u.UID = uid
	u.GID = gid

	// Try to set the home directory
	usr, err := user.LookupId(sudoUID)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"uid":   uid,
		}).Error("Failed to lookup SUDO_USER entry")
		return false
	}

	// Now store the home directory for that user
	u.HomeDir = usr.HomeDir
	u.Username = usr.Username

	return true
}

// SetFromCurrent will set the UserInfo details from the current user
func (u *UserInfo) SetFromCurrent() {
	u.UID = os.Getuid()
	u.GID = os.Getgid()

	if usr, err := user.Current(); err != nil {
		u.HomeDir = usr.HomeDir
		u.Username = usr.Username
	} else {
		log.WithFields(log.Fields{
			"error": err,
			"uid":   u.UID,
		}).Error("Failed to lookup current user")
		u.Username = os.Getenv("USERNAME")
		u.HomeDir = filepath.Join("/home", u.Username)
	}
}

// SetFromPackager will set the username/email fields from the legacy solus
// packager file.
func (u *UserInfo) SetFromPackager() bool {
	candidatePaths := []string{
		filepath.Join(u.HomeDir, ".solus", "packager"),
		filepath.Join(u.HomeDir, ".evolveos", "packager"),
	}

	// Attempt to parse one of the packager files
	for _, p := range candidatePaths {
		if !PathExists(p) {
			continue
		}
		cfg, err := ini.Load(p)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"path":  p,
			}).Error("Error loading INI file")
			continue
		}

		section, err := cfg.GetSection("Packager")
		if err != nil {
			log.WithFields(log.Fields{
				"path": p,
			}).Error("Missing [Packager] section in file")
			continue
		}

		uname, err := section.GetKey("Name")
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"path":  p,
			}).Error("Packager file has missing Name")
			continue
		}
		email, err := section.GetKey("Email")
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"path":  p,
			}).Error("Packager file has missing Email")
			continue
		}
		u.Name = uname.String()
		u.Email = email.String()
		return true
	}

	return false
}

// SetFromGit will set the username/email fields from the git config file
func (u *UserInfo) SetFromGit() bool {
	gitConfPath := filepath.Join(u.HomeDir, ".gitconfig")
	if !PathExists(gitConfPath) {
		return false
	}

	cfg, err := ini.Load(gitConfPath)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  gitConfPath,
		}).Error("Error loading gitconfig")
		return false
	}

	section, err := cfg.GetSection("user")
	if err != nil {
		log.WithFields(log.Fields{
			"path": gitConfPath,
		}).Error("Missing [user] section in gitconfig")
		return false
	}

	uname, err := section.GetKey("name")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  gitConfPath,
		}).Error("gitconfig file has missing name")
		return false
	}
	email, err := section.GetKey("email")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  gitConfPath,
		}).Error("gitconfig file has missing email")
		return false
	}
	u.Name = uname.String()
	u.Email = email.String()

	return true
}

// GetUserInfo will always succeed, as it will use a fallback policy until it
// finally comes up with a valid combination of name/email to use.
func GetUserInfo() *UserInfo {
	uinfo := &UserInfo{}

	// First up try to set the uid/gid
	if !uinfo.SetFromSudo() {
		uinfo.SetFromCurrent()
	}

	attempts := []func() bool{
		uinfo.SetFromPackager,
		uinfo.SetFromGit,
	}

	for _, a := range attempts {
		if a() {
			return uinfo
		}
	}

	uinfo.Name = FallbackUserName
	uinfo.Email = FallbackUserEmail

	return uinfo
}
