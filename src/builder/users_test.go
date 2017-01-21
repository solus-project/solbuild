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
	"strings"
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

	derp, foundDerp := pwd.Users["derpmcderpface"]
	if !foundDerp {
		t.Fatalf("Failed to find known user")
	}
	if derp.UID != 1001 {
		t.Fatalf("User ID wrong: %d vs expected 1001", derp.UID)
	}
	if derp.GID != 1002 {
		t.Fatalf("User GID wrong: %d vs expected 1002", derp.UID)
	}
	if derp.Home != "/home/derpmcderpface" {
		t.Fatalf("Wrong homedir: '%s' vs expected /home/derpmcderpface", derp.Home)
	}

	if shell := pwd.Users["root"].Shell; shell != "/bin/bash" {
		t.Fatalf("Wrong shell for root: %s", shell)
	}

	sudo, foundSudo := pwd.Groups["sudo"]
	if !foundSudo {
		t.Fatalf("I am without sudo")
	}
	if len(sudo.Members) != 2 {
		t.Fatalf("sudo has wrong member count of %d vs expected 2", len(sudo.Members))
	}
	members := strings.Join(sudo.Members, ",")
	if members != "ikey,derpmcderpface" {
		t.Fatalf("Wrong members for sudo: %s", members)
	}

	lightdm := pwd.Groups["lightdm"]
	if lightdm.ID != pwd.Users["lightdm"].GID {
		t.Fatalf("Wrong GID for lightdm: %d vs expected %d", lightdm.ID, pwd.Users["lightdm"].GID)
	}

	if len(lightdm.Members) != 0 {
		t.Fatalf("Myseriously have members of lightdm: |%s| %d", strings.Join(lightdm.Members, ", "), len(lightdm.Members))
	}
}
