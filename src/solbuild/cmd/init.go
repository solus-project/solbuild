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
	"github.com/solus-project/libosdev/commands"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var initCmd = &cobra.Command{
	Use:   "init [profile]",
	Short: "initialise a solbuild profile",
	Long: `Initialise a solbuild profile so that it can be used for subsequent
builds`,
	Run: initProfile,
}

func init() {
	initCmd.Flags().StringVarP(&profile, "profile", "p", DefaultProfile, "Build profile to use")
	RootCmd.AddCommand(initCmd)
}

func initProfile(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		profile = strings.TrimSpace(args[0])
	}
	bk := builder.NewBackingImage(profile)
	if bk.IsInstalled() {
		fmt.Printf("'%v' has already been initialised\n", profile)
		return
	}
	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "You must be root to run init profiles\n")
		os.Exit(1)
	}

	imgDir := builder.ImagesDir

	// Ensure directories exist
	if !builder.PathExists(imgDir) {
		if err := os.MkdirAll(imgDir, 00755); err != nil {
			log.WithFields(log.Fields{
				"dir":   imgDir,
				"error": err,
			}).Error("Failed to create images directory")
			os.Exit(1)
		}
		log.WithFields(log.Fields{
			"dir": imgDir,
		}).Debug("Created images directory")
	}

	// Now ensure we actually have said image
	if !bk.IsFetched() {
		com := []string{"-o", bk.ImagePathXZ, "-L", "--progress-bar", bk.ImageURI}
		log.WithFields(log.Fields{
			"uri": bk.ImageURI,
		}).Info("Fetching backing image")
		if err := commands.ExecStdoutArgs("curl", com); err != nil {
			log.WithFields(log.Fields{
				"uri":   bk.ImageURI,
				"error": err,
			}).Error("Failed to fetch image")
		}
	}

	fmt.Fprintf(os.Stderr, "Yay initialising for %v..\n", profile)
}
