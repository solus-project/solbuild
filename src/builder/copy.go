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
	// We'll need this one later
	_ "github.com/Sirupsen/logrus"
)

// CopyAll will copy the source asset into the given destdir.
// If the source is a directory, it will be recursively copied
// into the directory destdir.
//
// Note that all directories are created as 00755, as solbuild
// has no interest in the individual folder permissions, just
// the files themselves.
func CopyAll(source, destdir string) error {
	return errors.New("Not yet implemented")
}
