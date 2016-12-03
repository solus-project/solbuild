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

// A Profile is a configuration defining what backing image to use, what repos
// to add, etc.
type Profile struct {
	Name  string `toml:"-"`     // Name of this profile, set by file name not toml
	Image string `toml:"image"` // The backing image for this profile
}

var (
	// ProfilePaths is a set of locations for valid solbuild configuration files
	ProfilePaths = []string{
		"/etc/solbuild",
		"/usr/share/solbuild",
	}
)

// NewProfile will attempt to load the profile from the system paths
func NewProfile(name string) (*Profile, error) {
	return nil, errors.New("Not yet implemented")
}
