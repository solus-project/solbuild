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
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var buildCmd = &cobra.Command{
	Use:   "build [package.yml|pspec.xml]",
	Short: "build a package",
	Long: `Build the given package in a chroot environment, and upon success,
store those packages in the current directory`,
	RunE: buildPackage,
}

var tmpfs bool
var tmpfsSize string

func init() {
	buildCmd.Flags().StringVarP(&profile, "profile", "p", builder.DefaultProfile, "Build profile to use")
	buildCmd.Flags().BoolVarP(&tmpfs, "tmpfs", "t", false, "Enable building in a tmpfs")
	buildCmd.Flags().StringVarP(&tmpfsSize, "memory", "m", "", "Set the tmpfs size to use")
	buildCmd.Flags().BoolVarP(&CLIDebug, "debug", "d", false, "Enable debug messages")
	RootCmd.AddCommand(buildCmd)
}

func buildPackage(cmd *cobra.Command, args []string) error {
	pkgPath := ""

	if CLIDebug {
		log.SetLevel(log.DebugLevel)
	}

	if len(args) == 1 {
		pkgPath = args[0]
	} else {
		// Try to find the logical path..
		pkgPath = FindLikelyArg()
	}

	// Initialise the build manager
	manager, err := builder.NewManager()
	if err != nil {
		return nil
	}
	if err := manager.SetProfile(profile); err != nil {
		if err == builder.ErrInvalidProfile {
			builder.EmitProfileError(profile)
		}
		return nil
	}

	pkgPath = strings.TrimSpace(pkgPath)

	if pkgPath == "" {
		return errors.New("Require a filename to build")
	}

	pkg, err := builder.NewPackage(pkgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load package: %v\n", err)
		return nil
	}

	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "You must be root to run build packages\n")
		os.Exit(1)
	}

	// Set the package
	if err := manager.SetPackage(pkg); err != nil {
		if err == builder.ErrProfileNotInstalled {
			fmt.Fprintf(os.Stderr, "%v: Did you forget to init?\n", err)
		}
		return nil
	}

	manager.SetTmpfs(tmpfs, tmpfsSize)
	if err := manager.Build(); err != nil {
		log.Error("Failed to build packages")
		return nil
	}

	log.Info("Building succeeded")
	return nil
}
