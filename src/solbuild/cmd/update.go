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

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update a solbuild profile",
	Long: `Update the base image of the specified solbuild profile, helping to
minimize the build times in future updates with this profile.`,
	Run: updateProfile,
}

func init() {
	updateCmd.Flags().StringVarP(&profile, "profile", "p", "", "Build profile to use")
	updateCmd.Flags().BoolVarP(&CLIDebug, "debug", "d", false, "Enable debug messages")
	RootCmd.AddCommand(updateCmd)
}

func updateProfile(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		profile = strings.TrimSpace(args[0])
	}

	if CLIDebug {
		log.SetLevel(log.DebugLevel)
	}

	// Initialise the build manager
	manager, err := builder.NewManager()
	if err != nil {
		return
	}
	// Safety first..
	if err = manager.SetProfile(profile); err != nil {
		if err == builder.ErrProfileNotInstalled {
			fmt.Fprintf(os.Stderr, "%v: Did you forget to init?\n", err)
		}
		return
	}

	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "You must be root to run init profiles\n")
		os.Exit(1)
	}

	if err := manager.Update(); err != nil {
		if err == builder.ErrProfileNotInstalled {
			fmt.Fprintf(os.Stderr, "%v: Did you forget to init?\n", err)
		}
		os.Exit(1)
	}
}
