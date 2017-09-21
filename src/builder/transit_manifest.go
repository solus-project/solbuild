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
	"bytes"
	"errors"
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
func NewTransitManifest(target string) *TransitManifest {
	ret := &TransitManifest{}
	ret.Manifest.Version = "1.0"
	ret.Manifest.Target = target
	return ret
}

// AddFile will attempt to add a file to the payload for this package
func (t *TransitManifest) AddFile(path string) error {
	if !strings.HasSuffix(path, ".eopkg") {
		return ErrIllegalUpload
	}
	hash, err := FileSha256sum(path)
	if err != nil {
		return err
	}

	t.File = append(t.File, TransitManifestFile{
		Path:   filepath.Base(path),
		Sha256: hash,
	})
	return nil
}

// Write will dump the manifest to the given file path
func (t *TransitManifest) Write(path string) error {
	blob := bytes.Buffer{}
	tmenc := toml.NewEncoder(&blob)
	// Waste of bytes.
	tmenc.Indent = ""
	if err := tmenc.Encode(t); err != nil {
		return err
	}
	return ioutil.WriteFile(path, blob.Bytes(), 00644)
}
