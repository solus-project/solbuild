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

func TestPasswd(t *testing.T) {
	if _, err := NewPasswd("./@W'el@@"); err == nil {
		t.Fatalf("Should not be able to parse non existent file")
	}

	pwd, err := NewPasswd("testdata")
	if err != nil {
		t.Fatalf("Unable to parse known good passwd data: %v", err)
	}
	if len(pwd.Users) != 19 {
		t.Fatalf("Invalid number of users parsed: %v vs expected 19", len(pwd.Users))
	}
	if len(pwd.Groups) != 48 {
		t.Fatalf("Invalid number of groups parsed: %v vs expected 48", len(pwd.Groups))
	}
}
