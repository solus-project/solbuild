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
	log "github.com/Sirupsen/logrus"
	"os"
	"path/filepath"
)

const (
	// BindRepoDir is where we make repos available from the host side
	BindRepoDir = "/hostRepos"
)

// addLocalRepo will try to add the repo and bind mount it into the target
func (p *Package) addLocalRepo(o *Overlay, pkgManager *EopkgManager, repo *Repo) error {
	// Ensure the source exists too. Sorta helpful like that.
	if !PathExists(repo.URI) {
		return fmt.Errorf("Local repo does not exist!")
	}

	// Ensure the target mountpoint actually exists ...
	tgt := filepath.Join(o.MountPoint, BindRepoDir, repo.Name)
	if !PathExists(tgt) {
		if err := os.MkdirAll(tgt, 00755); err != nil {
			return err
		}
	}

	// for eopkg the repo index is eopkg-index.xml.xz..
	// tgtIndex := filepath.Join(tgt, "eopkg-index.xml.xz")

	return ErrNotImplemented
}

func (p *Package) removeRepos(pkgManager *EopkgManager, repos []string) error {
	if len(repos) < 1 {
		return nil
	}
	for _, id := range repos {
		log.WithFields(log.Fields{
			"name": id,
		}).Info("Removing repository")
		if err := pkgManager.RemoveRepo(id); err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"name":  id,
			}).Error("Failed to remove repository")
			return err
		}
	}
	return nil
}

// addRepos will add the specified filtered set of repos to the rootfs
func (p *Package) addRepos(o *Overlay, pkgManager *EopkgManager, repos []*Repo) error {
	if len(repos) < 1 {
		return nil
	}
	for _, repo := range repos {
		if repo.Local {
			log.WithFields(log.Fields{
				"name": repo.Name,
				"path": repo.URI,
			}).Info("Adding local repo to system")

			if err := p.addLocalRepo(o, pkgManager, repo); err != nil {
				log.WithFields(log.Fields{
					"name":  repo.Name,
					"error": err,
				}).Error("Failed to add local repo to system")
				return err
			}
			continue
		}
		log.WithFields(log.Fields{
			"name": repo.Name,
			"url":  repo.URI,
		}).Info("Adding repo to system")
		if err := pkgManager.AddRepo(repo.Name, repo.URI); err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"name":  repo.Name,
			}).Error("Failed to add repo to system")
			return err
		}
	}
	return nil
}

// ConfigureRepos will attempt to configure the repos according to the configuration
// of the manager.
func (p *Package) ConfigureRepos(o *Overlay, pkgManager *EopkgManager, profile *Profile) error {
	repos, err := pkgManager.GetRepos()
	if err != nil {
		return err
	}

	var removals []string

	// Find out which repos to remove
	if len(profile.RemoveRepos) == 1 && profile.RemoveRepos[0] == "*" {
		for _, r := range repos {
			removals = append(removals, r.ID)
		}
	} else {
		for _, r := range profile.RemoveRepos {
			removals = append(removals, r)
		}
	}

	if err := p.removeRepos(pkgManager, removals); err != nil {
		return err
	}

	var addRepos []*Repo

	if (len(profile.AddRepos) == 1 && profile.AddRepos[0] == "*") || len(profile.AddRepos) == 0 {
		for _, repo := range profile.Repos {
			addRepos = append(addRepos, repo)
		}
	} else {
		for _, id := range profile.AddRepos {
			addRepos = append(addRepos, profile.Repos[id])
		}
	}

	return p.addRepos(o, pkgManager, addRepos)
}
