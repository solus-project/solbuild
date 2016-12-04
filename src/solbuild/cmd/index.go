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

package cmd

import (
	"builder"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var indexCmd = &cobra.Command{
	Use:   "index [directory]",
	Short: "create repo index in the given directory",
	Long: `Use the given build profile to construct a repository index in the
given directory. If a directory is not specified, then the current directory
is used. This directory will be mounted inside the container and the Solus
machinery will be used to create a repository.`,
	RunE: indexPackages,
}

func init() {
	indexCmd.Flags().StringVarP(&profile, "profile", "p", "", "Build profile to use")
	indexCmd.Flags().BoolVarP(&CLIDebug, "debug", "d", false, "Enable debug messages")
	RootCmd.AddCommand(indexCmd)
}

func indexPackages(cmd *cobra.Command, args []string) error {
	if CLIDebug {
		log.SetLevel(log.DebugLevel)
	}

	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "You must be root to use index\n")
		os.Exit(1)
	}

	indexDir := "."
	if len(args) == 1 {
		indexDir = args[0]
	}

	indexDir = strings.TrimSpace(indexDir)

	// Initialise the build manager
	manager, err := builder.NewManager()
	if err != nil {
		return nil
	}
	// Safety first..
	if err = manager.SetProfile(profile); err != nil {
		return nil
	}

	// Set the package
	if err := manager.SetPackage(&builder.IndexPackage); err != nil {
		if err == builder.ErrProfileNotInstalled {
			fmt.Fprintf(os.Stderr, "%v: Did you forget to init?\n", err)
		}
		return nil
	}

	if err := manager.Index(indexDir); err != nil {
		log.Error("Index failure")
		return nil
	}

	log.Info("Indexing complete")
	return nil
}
