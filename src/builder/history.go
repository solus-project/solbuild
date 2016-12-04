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
	"time"
)

const (
	// MaxChangelogEntries is the absolute maximum number of entries we'll
	// parse and provide changelog entries for.
	MaxChangelogEntries = 10
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

// A PackageUpdate is a point in history in the git changes, which is parsed
// from a git.Commit
type PackageUpdate struct {
	Tag         string    // The associated git tag
	Author      string    // The author name of the change
	AuthorEmail string    // The author email of the change
	Body        string    // The associated message of the commit
	Time        time.Time // When the update took place
	ObjectID    string    // OID stored in string form
	Package     *Package  // Associated parsed package
}

// NewPackageUpdate will attempt to parse the given commit and provide a usable
// entry for the PackageHistory
func NewPackageUpdate(tag string, commit *git.Commit, objectID string) *PackageUpdate {
	signature := commit.Author()
	update := &PackageUpdate{Tag: tag}

	// We duplicate. cgo makes life difficult.
	update.Author = signature.Name
	update.AuthorEmail = signature.Email
	update.Body = commit.Message()
	update.Time = signature.When
	update.ObjectID = objectID

	return update
}

// CatGitBlob will return the contents of the given entry
func CatGitBlob(repo *git.Repository, entry *git.TreeEntry) ([]byte, error) {
	obj, err := repo.Lookup(entry.Id)
	if err != nil {
		return nil, err
	}
	blob, err := obj.AsBlob()
	if err != nil {
		return nil, err
	}
	return blob.Contents(), nil
}

// GetFileContents will attempt to read the entire object at path from
// the given tag, within that repo.
func GetFileContents(repo *git.Repository, tag, path string) ([]byte, error) {
	oid, err := git.NewOid(tag)
	if err != nil {
		return nil, err
	}
	commit, err := repo.Lookup(oid)
	if err != nil {
		return nil, err
	}
	treeObj, err := commit.Peel(git.ObjectTree)
	if err != nil {
		return nil, err
	}
	tree, err := treeObj.AsTree()
	if err != nil {
		return nil, err
	}
	entry, err := tree.EntryByPath(path)
	if err != nil {
		return nil, err
	}

	return CatGitBlob(repo, entry)
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

	updates := make(map[string]*PackageUpdate)

	// Iterate all of the tags
	err = repo.Tags.Foreach(func(name string, id *git.Oid) error {
		if name == "" || id == nil {
			return nil
		}

		var commit *git.Commit

		obj, err := repo.Lookup(id)
		if err != nil {
			return err
		}

		switch obj.Type() {
		// Unannotated tag
		case git.ObjectCommit:
			commit, err = obj.AsCommit()
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
			commit, err = repo.LookupCommit(tag.TargetId())
			if err != nil {
				return err
			}
			tags = append(tags, name)
		default:
			return fmt.Errorf("Internal git error, found %s", obj.Type().String())
		}
		if commit == nil {
			return nil
		}
		commitObj := NewPackageUpdate(name, commit, id.String())
		updates[name] = commitObj
		return nil
	})
	// Foreach went bork
	if err != nil {
		return nil, err
	}
	// Sort the tags by -refname
	sort.Sort(sort.Reverse(sort.StringSlice(tags)))

	numEntries := 0

	// Iterate the commit set in order
	for _, tagID := range tags {
		if numEntries > MaxChangelogEntries {
			break
		}
		update := updates[tagID]
		if update == nil {
			continue
		}
		b, err := GetFileContents(repo, update.ObjectID, "package.yml")
		if err != nil {
			return nil, err
		}

		var pkg *Package
		// Shouldn't *actually* bail here. Malformed packages do happen
		if pkg, err = NewYmlPackageFromBytes(b); err != nil {
			return nil, err
		}
		update.Package = pkg
		numEntries++
	}

	return nil, ErrNotImplemented
}
