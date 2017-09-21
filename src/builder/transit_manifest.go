//
// Copyright © 2017 Solus Project
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
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	// TransitManifestSuffix is the extension that a valid transit manifest must have
	TransitManifestSuffix = ".tram"
)

var (
	// ErrInvalidHeader will be returned when the [manifest] section is malformed
	ErrInvalidHeader = errors.New("Manifest contains an invalid header")

	// ErrMissingTarget will be returned when the target is not present
	ErrMissingTarget = errors.New("Manifest contains no target")

	// ErrMissingPayload will be returned when the [[file]]s are missing
	ErrMissingPayload = errors.New("Manifest does not contain a payload")

	// ErrInvalidPayload will be returned when the payload is in some way invalid
	ErrInvalidPayload = errors.New("Manifest contains an invalid payload")

	// ErrIllegalUpload is returned when someone is a spanner and tries uploading an unsupported file
	ErrIllegalUpload = errors.New("The manifest file is NOT an eopkg")
)

// A TransitManifest is provided by build servers to validate the upload of
// packages into the incoming directory.
//
// This is to ensure all uploads are intentional, complete and verifiable.
type TransitManifest struct {

	// Every .tram file has a [manifest] header - this will never change and is
	// version agnostic.
	Manifest struct {

		// Versioning to protect against future format changes
		Version string `toml:"version"`

		// The repo that the uploader is intending to upload *to*
		Target string `toml:"target"`
	}

	// A list of files that accompanied this .tram upload
	File []TransitManifestFile `toml:"file"`

	path string // Privately held path to the file
	dir  string // Where the .tram was loaded from
	id   string // Effectively our basename
}

// ID will return the unique ID for the transit manifest file
func (t *TransitManifest) ID() string {
	return t.id
}

// GetPaths will return the package paths as a slice of strings
func (t *TransitManifest) GetPaths() []string {
	var ret []string
	for i := range t.File {
		f := &t.File[i]
		ret = append(ret, filepath.Join(t.dir, f.Path))
	}
	return ret
}

// TransitManifestFile provides simple verification data for each file in the
// uploaded payload.
type TransitManifestFile struct {

	// Relative filename, i.e. nano-2.7.5-68-1-x86_64.eopkg
	Path string `toml:"path"`

	// Cryptographic checksum to allow integrity checks post-upload/pre-merge
	Sha256 string `toml:"sha256"`
}

// NewTransitManifest will attempt to load the transit manifest from the
// named path and perform *basic* validation.
func NewTransitManifest(path string) (*TransitManifest, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	ret := &TransitManifest{
		path: abs,
		dir:  filepath.Dir(abs),
		id:   filepath.Base(abs),
	}

	blob, err := ioutil.ReadFile(ret.path)
	if err != nil {
		return nil, err
	}

	if _, err := toml.Decode(string(blob), ret); err != nil {
		return nil, err
	}

	ret.Manifest.Target = strings.TrimSpace(ret.Manifest.Target)
	ret.Manifest.Version = strings.TrimSpace(ret.Manifest.Version)

	if ret.Manifest.Version != "1.0" {
		return nil, ErrInvalidHeader
	}

	if len(ret.Manifest.Target) < 1 {
		return nil, ErrMissingTarget
	}

	if len(ret.File) < 1 {
		return nil, ErrMissingPayload
	}

	for i := range ret.File {
		f := &ret.File[i]
		f.Path = strings.TrimSpace(f.Path)
		f.Sha256 = strings.TrimSpace(f.Sha256)

		if len(f.Path) < 1 || len(f.Sha256) < 1 {
			return nil, ErrInvalidPayload
		}

		if !strings.HasSuffix(f.Path, ".eopkg") {
			return nil, ErrIllegalUpload
		}
	}

	return ret, nil
}

// ValidatePayload will verify the files listed in the manifest locally, ensuring
// that they actually exist, and that the hashes match to prevent any corrupted
// uploads being inadvertently imported
func (t *TransitManifest) ValidatePayload() error {
	for i := range t.File {
		f := &t.File[i]
		path := filepath.Join(t.dir, f.Path)
		sha, err := FileSha256sum(path)
		if err != nil {
			return err
		}
		if sha != f.Sha256 {
			return fmt.Errorf("Invalid SHA256 for '%s'. Local: '%s'", f.Path, sha)
		}
	}
	return nil
}
