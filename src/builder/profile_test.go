//
// Copyright © 2016-2017 Ikey Doherty <ikey@solus-project.com>
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
	"testing"
)

const (
	ProfileTestFile = "testdata/unstable.profile"
)

func TestLoadProfile(t *testing.T) {
	if _, err := NewProfileFromPath("@'werlq;krqr8u3283"); err == nil {
		t.Fatal("Loaded a file that doesn't exist!")
	}

	profile, err := NewProfileFromPath(ProfileTestFile)
	if err != nil {
		t.Fatalf("Failed to load configuration from valid path: %v", err)
	}
	if profile == nil {
		t.Fatal("No error but nil profile")
	}
	if profile.Image != "unstable-x86_64" {
		t.Fatalf("Wrong image in profile: %v", profile.Image)
	}
	if repo, ok := profile.Repos["Solus"]; ok {
		if repo.URI != "https://packages.solus-project.com/unstable/eopkg-index.xml.xz" {
			t.Fatalf("Wrong Solus URI: %v", repo.URI)
		}
	} else {
		t.Fatal("Missing Solus repo")
	}
	if _, ok := profile.Repos["Bob"]; ok {
		t.Fatal("Should not have a repo here!")
	}
	if len(profile.RemoveRepos) != 0 {
		t.Fatalf("Invalid number of remove repos: %d", len(profile.RemoveRepos))
	}
	if len(profile.AddRepos) != 1 {
		t.Fatalf("Invalid number of add repos: %d", len(profile.AddRepos))
	}
	if len(profile.Repos) != 3 {
		t.Fatalf("Invalid number of repos: %d", len(profile.Repos))
	}
	if profile.AddRepos[0] != "Solus" {
		t.Fatalf("Invalid AddRepos: %s", profile.AddRepos[0])
	}
}
