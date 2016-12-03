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
	"testing"
)

const (
	ProfileTestFile = "testdata/unstable.profile"
)

func TestLoadProfile(t *testing.T) {
	if _, err := NewProfileFromPath("@'werlq;krqr8u3283"); err == nil {
		t.Fatal("Loaded a file that doesn't exist!")
	}

	_, err := NewProfileFromPath(ProfileTestFile)
	if err != nil {
		t.Fatalf("Failed to load configuration from valid path")
	}
}
