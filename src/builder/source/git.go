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
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/libgit2/git2go"
	"github.com/solus-project/libosdev/commands"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	// GitSourceDir is the base directory for all cached git sources
	GitSourceDir = "/var/lib/solbuild/sources/git"
)

var (
	// ErrGitNoContinue is returned when git processing cannot continue
	ErrGitNoContinue = errors.New("Fatal errors in git fetch")
)

// A GitSource as referenced by `ypkg` build spec. A git source must have
// a valid ref to check out to.
type GitSource struct {
	URI       string
	Ref       string
	BaseName  string
	ClonePath string // This is where we will have cloned into
}

// NewGit will create a new GitSource for the given URI & ref combination.
func NewGit(uri, ref string) (*GitSource, error) {
	// Ensure we have a valid URL first.
	urlObj, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	bs := filepath.Base(urlObj.Path)
	if !strings.HasSuffix(bs, ".git") {
		bs += ".git"
	}

	// This is where we intend to clone to locally
	clonePath := filepath.Join(GitSourceDir, urlObj.Host, filepath.Dir(urlObj.Path), bs)

	g := &GitSource{
		URI:       uri,
		Ref:       ref,
		BaseName:  bs,
		ClonePath: clonePath,
	}

	return g, nil
}

// completed is called when the fetch is done
func (g *GitSource) completed(r git.RemoteCompletion) git.ErrorCode {
	log.WithFields(log.Fields{
		"source": g.BaseName,
	}).Info("Completed fetch of git source")
	return 0
}

// message will be called to emit standard git text to the terminal
func (g *GitSource) message(str string) git.ErrorCode {
	os.Stdout.Write([]byte(str))
	return 0
}

// CreateCallbacks will create the default git callbacks
func (g *GitSource) CreateCallbacks() git.RemoteCallbacks {
	return git.RemoteCallbacks{
		SidebandProgressCallback: g.message,
	}
}

// Clone will set do a bare mirror clone of the remote repo to the local
// cache.
func (g *GitSource) Clone() error {
	// Attempt cloning
	log.WithFields(log.Fields{
		"uri": g.URI,
	}).Info("Cloning git source")

	fetchOpts := &git.FetchOptions{
		RemoteCallbacks: g.CreateCallbacks(),
	}

	_, err := git.Clone(g.URI, g.ClonePath, &git.CloneOptions{
		Bare:         false,
		FetchOptions: fetchOpts,
	})
	return err
}

// HasTag will attempt to find the tag, if possible
func (g *GitSource) HasTag(repo *git.Repository, tagName string) bool {
	haveTag := false
	repo.Tags.Foreach(func(name string, id *git.Oid) error {
		if name == "refs/tags/"+tagName {
			haveTag = true
		}
		return nil
	})
	return haveTag
}

// fetch will attempt
func (g *GitSource) fetch(repo *git.Repository) error {
	log.WithFields(log.Fields{
		"uri": g.URI,
	}).Info("Git fetching existing clone")
	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"uri":   g.URI,
		}).Error("Failed to find git remote")
		return err
	}

	fetchOpts := &git.FetchOptions{
		RemoteCallbacks: g.CreateCallbacks(),
	}
	if err := remote.Fetch([]string{}, fetchOpts, ""); err != nil {
		return err
	}
	return nil
}

// GetCommitID will attempt to find the oid of the selected ref type
func (g *GitSource) GetCommitID(repo *git.Repository) string {
	oid := ""
	// Attempt to find the branch
	branch, err := repo.LookupBranch(g.Ref, git.BranchAll)
	if err == nil {
		oid = branch.Target().String()
		log.WithFields(log.Fields{
			"branch": g.Ref,
			"sha":    oid,
		}).Info("Found git commit of branch")
		return oid
	}

	tagName := g.Ref
	if !strings.HasPrefix(tagName, "refs/tags") {
		tagName = "refs/tags/" + tagName
	}

	repo.Tags.Foreach(func(name string, id *git.Oid) error {
		if name == tagName {
			oid = id.String()
			// Force break the foreach
			return errors.New("")
		}
		return nil
	})

	// Tag set the oid
	if oid != "" {
		log.WithFields(log.Fields{
			"tag": tagName,
			"sha": oid,
		}).Info("Found git commit of tag")
		return oid
	}

	// Check the oid is valid
	oid = g.Ref
	obj, err := git.NewOid(oid)
	if err != nil {
		return ""
	}

	// Check if its a commit
	_, err = repo.Lookup(obj)
	if err != nil {
		return ""
	}
	log.WithFields(log.Fields{
		"tag": tagName,
		"sha": oid,
	}).Info("Found git commit")
	return obj.String()
}

// GetHead will attempt to gain the OID for head
func (g *GitSource) GetHead(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", err
	}
	return head.Target().String(), nil
}

// resetOnto will attempt to reset the repo (hard) onto the given commit
func (g *GitSource) resetOnto(repo *git.Repository, ref string) error {
	// this stuff _really_ shouldn't happen but oh well.
	oid, err := git.NewOid(ref)
	if err != nil {
		return err
	}
	commitFind, err := repo.Lookup(oid)
	if err != nil {
		return err
	}

	commitObj, err := commitFind.Peel(git.ObjectCommit)
	if err != nil {
		return err
	}
	commit, err := commitObj.AsCommit()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"sha": ref,
	}).Info("Resetting git repository to commit")

	checkOpts := &git.CheckoutOpts{
		Strategy: git.CheckoutForce | git.CheckoutRemoveUntracked | git.CheckoutRemoveIgnored}

	if err := repo.ResetToCommit(commit, git.ResetHard, checkOpts); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"sha":   ref,
		}).Error("Failed to reset git repository")
		return err
	}

	return nil
}

// submodules will handle setup of the git submodules after a
// reset has taken place.
func (g *GitSource) submodules() error {
	// IDK What else to tell ya, git2go submodules is broken
	cmd := []string{"submodule", "update", "--init"}
	return commands.ExecStdoutArgsDir(g.ClonePath, "git", cmd)
}

// Fetch will attempt to download the git tree locally. If it already exists
// then we'll make an attempt to update it.
func (g *GitSource) Fetch() error {
	hadRepo := true

	// First things first, clone if necessary
	if !PathExists(g.ClonePath) {
		if err := g.Clone(); err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"uri":   g.URI,
			}).Error("Failed to clone remote repository")
			return err
		}
		hadRepo = false
	}

	// Now open the repo and validate it
	repo, err := git.OpenRepository(g.ClonePath)
	if err != nil {
		return err
	}

	wantedCommit := g.GetCommitID(repo)
	if wantedCommit == "" {
		// Logic here being we just cloned it. Where is it?
		if !hadRepo {
			return fmt.Errorf("Cannot continue with git processing")
		}
		// So try to fetch it
		if err := g.fetch(repo); err != nil {
			return err
		}
		// Re-establish the wanted commit
		wantedCommit = g.GetCommitID(repo)
	}

	// Can't proceed now. Just doesn't exist
	if wantedCommit == "" {
		return ErrGitNoContinue
	}

	// Attempt reset
	if err := g.resetOnto(repo, wantedCommit); err != nil {
		return err
	}

	// Check out submodules
	if err := g.submodules(); err != nil {
		return err
	}

	return nil
}

// IsFetched will check if we have the ref available, if not it will return
// false so that Fetch() can do the hard work.
func (g *GitSource) IsFetched() bool {
	return false
}

// GetBindConfiguration will return a config that enables bind mounting
// the bare git clone from the host side into the container, at which
// point ypkg can git clone from the bare git into a new tree and check
// out, make changes, etc.
func (g *GitSource) GetBindConfiguration(sourcedir string) BindConfiguration {
	return BindConfiguration{
		g.ClonePath,
		filepath.Join(sourcedir, g.BaseName),
	}
}

// GetIdentifier will return a human readable string to represent this
// git source in the event of errors.
func (g *GitSource) GetIdentifier() string {
	return fmt.Sprintf("%s#%s", g.URI, g.Ref)
}
