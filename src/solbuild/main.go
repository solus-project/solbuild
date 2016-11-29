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

package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/solus-project/libosdev/commands"
	"os"
	"syscall"
)

func wgetTest(url string) error {
	defer func() {
		os.Remove("index.html")
	}()
	return commands.ExecStdoutArgs("wget", []string{url})
}

func main() {
	// Drop main namespace stuff
	log.Info("Entering new namespace for child processes")
	if err := syscall.Unshare(syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC); err != nil {
		panic(err)
	}

	log.Info("Attempting download (should work)")
	if err := wgetTest("https://google.com"); err != nil {
		fmt.Fprintf(os.Stderr, "Downloading should still work, but doesn't: %v\n", err)
		os.Exit(1)
	}

	log.Info("Dropping networking")
	// Drop networking now
	if err := syscall.Unshare(syscall.CLONE_NEWNET | syscall.CLONE_NEWUTS); err != nil {
		panic(err)
	}

	log.Info("Redownloading, should fail")
	if err := wgetTest("https://google.com"); err != nil {
		fmt.Fprintf(os.Stderr, "Great, networking doesn't work :)\n")
	} else {
		fmt.Fprintf(os.Stderr, "Networking should not be working!\n")
		os.Exit(1)
	}
}
