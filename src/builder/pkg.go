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
	"builder/source"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"os"
	"strings"
)

// PackageType is simply the type of package we're building, i.e. xml / pspec
type PackageType string

const (
	// PackageTypeXML is the legacy package format, to be removed with sol introduction.
	PackageTypeXML PackageType = "legacy"

	// PackageTypeYpkg is the native build format of Solus, the package.yml format
	PackageTypeYpkg PackageType = "ypkg"

	// PackageTypeIndex is a faux type to enable indexing
	PackageTypeIndex PackageType = "index"
)

var (
	// IndexPackage is used by the index command to make use of the overlayfs
	// system.
	IndexPackage = Package{
		Name:    "index",
		Version: "0.1",
		Type:    PackageTypeIndex,
		Release: 1,
		Path:    "",
	}
)

// Package is the main item we deal with, avoiding the internals
type Package struct {
	Name       string          // Name of the package
	Version    string          // Version of this package
	Release    int             // Solus upgrades are based entirely on relno
	Type       PackageType     // ypkg or pspec.xml legacy
	Path       string          // Path to the build spec
	Sources    []source.Source // Each package has 0 or more sources that we fetch
	CanNetwork bool            // Only applicable to ypkg builds
}

// YmlPackage is a parsed ypkg build file
type YmlPackage struct {
	Name       string
	Version    string
	Release    int
	Networking bool // If set to false (default) we disable networking in the build
	Source     []map[string]string
}

// XMLUpdate represents an update in the package history
type XMLUpdate struct {
	Release int `xml:"release,attr"`
	Date    string
	Version string
	Comment string
	Name    string
	Email   string
}

// XMLArchive is an <Archive> line in Source section
type XMLArchive struct {
	Type    string `xml:"type,attr"`
	SHA1Sum string `xml:"sha1sum,attr"`
	URI     string `xml:",chardata"`
}

// XMLSource is the actual source info for each pspec.xml
type XMLSource struct {
	Homepage string
	Name     string
	Archive  []XMLArchive
}

// XMLPackage contains all of the pspec.xml metadata
type XMLPackage struct {
	Name    string
	Source  XMLSource
	History []XMLUpdate `xml:"History>Update"`
}

// NewPackage will attempt to parse the given path, and return a new Package
// instance if this succeeds.
func NewPackage(path string) (*Package, error) {
	if strings.HasSuffix(path, ".xml") {
		return NewXMLPackage(path)
	}
	return NewYmlPackage(path)
}

// NewXMLPackage will attempt to parse the pspec.xml file @ path
func NewXMLPackage(path string) (*Package, error) {
	var by []byte
	var err error
	var fi *os.File

	fi, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	by, err = ioutil.ReadAll(fi)
	if err != nil {
		return nil, err
	}
	xpkg := &XMLPackage{}
	if err = xml.Unmarshal(by, xpkg); err != nil {
		return nil, err
	}
	if len(xpkg.History) < 1 {
		return nil, errors.New("xml: Malformed pspec file")
	}

	upd := xpkg.History[0]
	ret := &Package{
		Name:       strings.TrimSpace(xpkg.Source.Name),
		Version:    strings.TrimSpace(upd.Version),
		Release:    upd.Release,
		Type:       PackageTypeXML,
		Path:       path,
		CanNetwork: true,
	}

	for _, archive := range xpkg.Source.Archive {
		source, err := source.New(archive.URI, archive.SHA1Sum, true)
		if err != nil {
			return nil, err
		}
		ret.Sources = append(ret.Sources, source)
	}

	if ret.Name == "" {
		return nil, errors.New("xml: Missing name in package")
	}
	if ret.Version == "" {
		return nil, errors.New("xml: Missing version in package")
	}
	if ret.Release < 0 {
		return nil, fmt.Errorf("xml: Invalid release in package: %d", ret.Release)
	}
	return ret, nil
}

// NewYmlPackage will attempt to parse the ypkg package.yml file @ path
func NewYmlPackage(path string) (*Package, error) {
	var by []byte
	var err error
	var fi *os.File

	fi, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	by, err = ioutil.ReadAll(fi)
	if err != nil {
		return nil, err
	}
	ret, err := NewYmlPackageFromBytes(by)
	if err != nil {
		return nil, err
	}
	ret.Path = path
	return ret, nil
}

// NewYmlPackageFromBytes will attempt to parse the ypkg package.yml in memory
func NewYmlPackageFromBytes(by []byte) (*Package, error) {
	var err error

	ypkg := &YmlPackage{Networking: false}
	if err = yaml.Unmarshal(by, ypkg); err != nil {
		return nil, err
	}

	ret := &Package{
		Name:       strings.TrimSpace(ypkg.Name),
		Version:    strings.TrimSpace(ypkg.Version),
		Release:    ypkg.Release,
		Type:       PackageTypeYpkg,
		CanNetwork: ypkg.Networking,
	}

	for _, row := range ypkg.Source {
		for key, value := range row {
			source, err := source.New(key, value, false)
			if err != nil {
				return nil, err
			}
			ret.Sources = append(ret.Sources, source)
		}
	}

	if ret.Name == "" {
		return nil, errors.New("ypkg: Missing name in package")
	}
	if ret.Version == "" {
		return nil, errors.New("ypkg: Missing version in package")
	}
	if ret.Release < 0 {
		return nil, fmt.Errorf("ypkg: Invalid release in package: %d", ret.Release)
	}
	return ret, nil
}
