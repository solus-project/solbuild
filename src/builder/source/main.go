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
	"os"
)

const (
	// SourceDir is where we store all tarballs
	SourceDir = "/var/lib/solbuild/sources"

	// SourceStagingDir is where we initially fetch downloads
	SourceStagingDir = "/var/lib/solbuild/sources/staging"
)

// A BindConfiguration is used by a source as a way to express bind
// mounts required for a given source.
//
// In solbuild, *all* sources are bind mounted to the target cache,
// regardless of their type.
//
// Special care is taken to ensure that they will be bound in a way
// compatible with the target system.
type BindConfiguration struct {
    BindSource string // The localy cached source
    BindTarget string // Target within the filesystem
}

// A Source is a general representation of source listed in a package
// spec file.
//
// Source's may be of multiple types, but all are abstracted and dealt
// with by the interfaces.
type Source interface {

    // IsFetched is called during the early build process to determine
    // whether this source is available for use.
    IsFetched() bool

    // Fetch will attempt to fetch the this source locally and cache it.
    Fetch() error

    // GetBindConfiguration should return a valid configuration specifying
    // the origin on our local filesystem, and the target within the container.
    GetBindConfiguration(rootfs string) BindConfiguration
}

// PathExists is a helper function to determine the existence of a file path
func PathExists(path string) bool {
	if st, err := os.Stat(path); err == nil && st != nil {
		return true
	}
	return false
}
