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
	"fmt"
	"github.com/libgit2/git2go"
	"sort"
)

// PackageHistory is an automatic changelog generated from the changes to
// the package.yml file during the history of the package.
//
// Through this system, we provide a `history.xml` file to `ypkg-build`
// inside the container, which allows it to export the changelog back to
// the user.
//
// This provides a much more natural system than having dedicated changelog
// files in package gits, as it reduces any and all duplication.
// We also have the opportunity to parse natural elements from the git history
// to make determinations as to the update *type*, such as a security update,
// or an update that requires a reboot to the users system.
//
// Currently we're only scoping for security update notification, though
// more features will come in time.
type PackageHistory struct {
}

// NewPackageHistory will attempt to analyze the git history at the given
// repository path, and return a usable instance of PackageHistory for writing
// to the container history.xml file.
func NewPackageHistory(path string) (*PackageHistory, error) {
	repo, err := git.OpenRepository(path)
	if err != nil {
		return nil, err
	}
	// Get all the tags
	var tags []string
	tags, err = repo.Tags.List()
	if err != nil {
		return nil, err
	}

	// Iterate all of the tags
	err = repo.Tags.Foreach(func(name string, id *git.Oid) error {
		if name == "" || id == nil {
			return nil
		}

		obj, err := repo.Lookup(id)
		if err != nil {
			return err
		}

		switch obj.Type() {
		// Unannotated tag
		case git.ObjectCommit:
			_, err := obj.AsCommit()
			if err != nil {
				return err
			}
			tags = append(tags, name)
		// Annotated tag with commit target
		case git.ObjectTag:
			tag, err := obj.AsTag()
			if err != nil {
				return err
			}
			_, err = repo.LookupCommit(tag.TargetId())
			if err != nil {
				return err
			}
			tags = append(tags, name)
		default:
			return fmt.Errorf("Internal git error, found %s", obj.Type().String())
		}
		return nil
	})
	// Foreach went bork
	if err != nil {
		return nil, err
	}
	// Sort the tags by -refname
	sort.Sort(sort.Reverse(sort.StringSlice(tags)))

	// Iterate the commit set in order
	for _, tagID := range tags {
		fmt.Println(tagID)
	}

	return nil, ErrNotImplemented
}
