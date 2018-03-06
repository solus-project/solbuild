//
// Copyright © 2016-2017 Ikey Doherty <ikey@solus-project.com>
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
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/solus-project/libosdev/commands"
	"github.com/spf13/cobra"
)

var (
	initCmd = &cobra.Command{
		Use:   "init [profile]",
		Short: "initialise a solbuild profile",
		Long: `Initialise a solbuild profile so that it can be used for subsequent
builds`,
		Run: initProfile,
	}

	// Whether we should automatically update the image after initialising it.
	autoUpdate bool
)

func init() {
	initCmd.Flags().BoolVarP(&autoUpdate, "update", "u", false, "Automatically update the new image")
	RootCmd.AddCommand(initCmd)
}

func doInit(manager *builder.Manager) {
	prof := manager.GetProfile()
	bk := builder.NewBackingImage(prof.Image)
	if bk.IsInstalled() {
		fmt.Printf("'%v' has already been initialised\n", profile)
		return
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
		downloadImage(bk, false)
	}

	// Decompress the image
	log.WithFields(log.Fields{
		"source": bk.ImagePathXZ,
		"target": bk.ImagePath,
	}).Debug("Decompressing backing image")

	if err := commands.ExecStdoutArgsDir(builder.ImagesDir, "unxz", []string{bk.ImagePathXZ}); err != nil {
		log.WithFields(log.Fields{
			"source": bk.ImagePathXZ,
			"error":  err,
		}).Error("Failed to decompress image")
	}

	log.WithFields(log.Fields{
		"profile": profile,
	}).Info("Profile successfully initialised")
}

// Downloads an image using net/http.
func downloadImage(bk *builder.BackingImage, progressBar bool) (err error) {
	file, err := os.Create(bk.ImagePathXZ)
	if err != nil {
		log.WithFields(log.Fields{
			"path":  bk.ImagePathXZ,
			"error": err,
		}).Error("Failed to create file")
		return err
	}

	defer func() {
		if err != nil {
			os.Remove(bk.ImagePathXZ)
		}
	}()

	defer file.Close()

	resp, err := http.Get(bk.ImageURI)
	if err != nil {
		log.WithFields(log.Fields{
			"uri":   bk.ImageURI,
			"error": err,
		}).Error("Failed to fetch image")
		return err
	}

	defer resp.Body.Close()

	bytesRemaining := resp.ContentLength
	done := false
	buf := make([]byte, 32*1024)
	for !done {
		bytesRead, err := resp.Body.Read(buf)
		if err == io.EOF {
			done = true
		} else if err != nil {
			log.WithFields(log.Fields{
				"uri":   bk.ImageURI,
				"error": err,
			}).Error("Failed to fetch image")
			return err
		}

		_, err = file.Write(buf[:bytesRead])
		if err != nil {
			log.WithFields(log.Fields{
				"uri":   bk.ImagePathXZ,
				"error": err,
			}).Error("Failed to write to file")
			return err
		}

		bytesRemaining -= int64(bytesRead)
	}

	return nil
}

// doUpdate will perform an update to the image after the initial init stage
func doUpdate(manager *builder.Manager) {
	if err := manager.Update(); err != nil {
		os.Exit(1)
	}
}

func initProfile(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		profile = strings.TrimSpace(args[0])
	}

	if CLIDebug {
		log.SetLevel(log.DebugLevel)
	}
	log.StandardLogger().Formatter.(*log.TextFormatter).DisableColors = builder.DisableColors

	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "You must be root to run init profiles\n")
		os.Exit(1)
	}

	// Now we'll update the newly initialised image
	manager, err := builder.NewManager()
	if err != nil {
		return
	}

	// Safety first..
	if err = manager.SetProfile(profile); err != nil {
		return
	}

	doInit(manager)

	if autoUpdate {
		doUpdate(manager)
	}
}
