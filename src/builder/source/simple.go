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

package source

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	log "github.com/Sirupsen/logrus"
	"github.com/cheggaaa/pb"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// A SimpleSource is a tarball or other source for a package
type SimpleSource struct {
	URI  string
	File string // Basename of the file

	legacy    bool   // If this is ypkg or not
	validator string // Validation key for this source
}

// NewSimple will create a new source instance
func NewSimple(uri, validator string, legacy bool) (*SimpleSource, error) {
	// Ensure the URI is actually valid.
	uriObj, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	ret := &SimpleSource{
		URI:       uri,
		File:      filepath.Base(uriObj.Path),
		legacy:    legacy,
		validator: validator,
	}
	return ret, nil
}

// GetIdentifier will return the URI associated with this source.
func (s *SimpleSource) GetIdentifier() string {
	return s.URI
}

// GetBindConfiguration will return the pair for binding our tarballs.
func (s *SimpleSource) GetBindConfiguration(rootfs string) BindConfiguration {
	return BindConfiguration{
		BindSource: s.GetPath(s.validator),
		BindTarget: filepath.Join(rootfs, s.File),
	}
}

// GetPath gets the path on the filesystem of the source
func (s *SimpleSource) GetPath(hash string) string {
	return filepath.Join(SourceDir, hash, s.File)
}

// GetSHA1Sum will return the sha1sum for the given path
func (s *SimpleSource) GetSHA1Sum(path string) (string, error) {
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
func (s *SimpleSource) GetSHA256Sum(path string) (string, error) {
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
func (s *SimpleSource) IsFetched() bool {
	return PathExists(s.GetPath(s.validator))
}

// download will perform the actual download of the source in
// question.
//
// TODO: Use a progressbar dependent on whether we're in server
// mode or not.
func (s *SimpleSource) download(destination string) error {
	// Fix up the http client
	client := http.Client{
		Timeout: time.Second * 15,
	}
	// Grab a http request
	resp, err := client.Get(s.URI)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(destination)
	if err != nil {
		return err
	}

	defer out.Close()

	pbar := pb.New64(resp.ContentLength).Prefix(filepath.Base(destination))
	pbar.SetUnits(pb.U_BYTES)
	pbar.SetMaxWidth(80)
	pbar.ShowSpeed = true
	reader := pbar.NewProxyReader(resp.Body)
	pbar.Start()
	defer func() {
		pbar.Update()
		pbar.Finish()
	}()

	if _, err := io.Copy(out, reader); err != nil {
		return err
	}
	return nil
}

// Fetch will download the given source and cache it locally
func (s *SimpleSource) Fetch() error {
	// Now go and download it
	log.WithFields(log.Fields{
		"uri": s.URI,
	}).Debug("Downloading source")

	destPath := filepath.Join(SourceStagingDir, s.File)

	// Check staging is available
	if !PathExists(SourceStagingDir) {
		if err := os.MkdirAll(SourceStagingDir, 00755); err != nil {
			return err
		}
	}

	// Grab the file
	if err := s.download(destPath); err != nil {
		return err
	}

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
	dest := filepath.Join(tgtDir, s.File)
	if err := os.Rename(destPath, dest); err != nil {
		return err
	}
	// If the file has a sha1sum set, symlink it to the sha256sum because
	// it's a legacy archive (pspec.xml)
	if s.legacy {
		sha, err := s.GetSHA1Sum(dest)
		if err != nil {
			return err
		}
		tgtLink := filepath.Join(SourceDir, sha)
		if err := os.Symlink(hash, tgtLink); err != nil {
			return err
		}
	}
	return nil
}
