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
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

const (
	// DefaultProfile is the profile solbuild will use when none are specified
	DefaultProfile = "main"
)

var profile string

// RootCmd is the main entry point into solbuild
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build a package",
	Long: `Build the given package in a chroot environment, and upon success,
store those packages in the current directory`,
	Run: buildPackage,
}

func init() {
	buildCmd.Flags().StringVarP(&profile, "profile", "p", DefaultProfile, "Build profile to use")
	RootCmd.AddCommand(buildCmd)
}

func buildPackage(cmd *cobra.Command, args []string) {
	fmt.Fprintf(os.Stderr, "Yay building for %v..\n", profile)
}
