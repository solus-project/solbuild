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

var chrootCmd = &cobra.Command{
	Use:   "chroot [package.yml|pspec.xml]",
	Short: "chroot into package's build environment",
	Long: `Interactively chroot into the package's build environment, to enable
further inspection when issues aren't immediately resolvable, i.e. pkg-config
dependencies.`,
	RunE: chrootPackage,
}

func init() {
	chrootCmd.Flags().StringVarP(&profile, "profile", "p", builder.DefaultProfile, "Build profile to use")
	RootCmd.AddCommand(chrootCmd)
}

func chrootPackage(cmd *cobra.Command, args []string) error {
	pkgPath := ""

	if len(args) == 1 {
		pkgPath = args[0]
	} else {
		// Try to find the logical path..
		pkgPath = FindLikelyArg()
	}

	pkgPath = strings.TrimSpace(pkgPath)

	if pkgPath == "" {
		return errors.New("Require a filename to chroot")
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

	pkg, err := builder.NewPackage(pkgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load package: %v\n", err)
		return nil
	}

	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "You must be root to use chroot\n")
		os.Exit(1)
	}

	// Set the package
	if err := manager.SetPackage(pkg); err != nil {
		if err == builder.ErrProfileNotInstalled {
			fmt.Fprintf(os.Stderr, "%v: Did you forget to init?\n", err)
		}
		return nil
	}

	if err := manager.Chroot(); err != nil {
		log.Error("Chroot failure")
		return nil
	}

	log.Info("Chroot complete")
	return nil
}
