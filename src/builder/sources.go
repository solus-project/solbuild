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
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"github.com/solus-project/libosdev/commands"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	// SourceDir is where we store all tarballs
	SourceDir = "/var/lib/solbuild/sources"

	// SourceStagingDir is where we initially fetch downloads
	SourceStagingDir = "/var/lib/solbuild/sources/staging"
)

// A Source is a tarball or other source for a package
type Source struct {
	SHA1Sum   string
	SHA256Sum string
	URI       string
	File      string // Basename of the file
}

// NewSource will create a new source instance
func NewSource(uri string) *Source {
	// TODO: Use a better method than filepath here
	return &Source{
		URI:  uri,
		File: filepath.Base(uri),
	}
}

// GetPath gets the path on the filesystem of the source
func (s *Source) GetPath(hash string) string {
	return filepath.Join(SourceDir, hash, filepath.Base(s.URI))
}

// GetSHA1Sum will return the sha1sum for the given path
func (s *Source) GetSHA1Sum(path string) (string, error) {
	inp, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha1.New()
	hash.Write(inp)
	sum := hash.Sum(nil)
	return hex.EncodeToString(sum), nil
}

// GetSHA256Sum will return the sha1sum for the given path
func (s *Source) GetSHA256Sum(path string) (string, error) {
	inp, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	hash.Write(inp)
	sum := hash.Sum(nil)
	return hex.EncodeToString(sum), nil
}

// IsFetched will determine if the source is already present
func (s *Source) IsFetched(expectedHash string) bool {
	return PathExists(s.GetPath(expectedHash))
}

// Fetch will download the given source and cache it locally
func (s *Source) Fetch() error {
	base := filepath.Base(s.URI)

	destPath := filepath.Join(SourceStagingDir, base)

	// Check staging is available
	if !PathExists(SourceStagingDir) {
		if err := os.MkdirAll(SourceStagingDir, 00755); err != nil {
			return err
		}
	}

	// Download to staging
	command := []string{
		"-L",
		"-o",
		destPath,
		"--progress-bar",
		s.URI,
	}

	if err := commands.ExecStdoutArgs("curl", command); err != nil {
		return err
	}

	// TODO: Check if legacy or not..
	hash, err := s.GetSHA256Sum(destPath)
	if err != nil {
		return err
	}

	// Make the target directory
	tgtDir := filepath.Join(SourceDir, hash)
	if !PathExists(tgtDir) {
		if err := os.MkdirAll(tgtDir, 00755); err != nil {
			return err
		}
	}
	// Move from staging into hash based directory
	dest := filepath.Join(tgtDir, base)
	if err := os.Rename(destPath, dest); err != nil {
		return err
	}
	return nil
}
