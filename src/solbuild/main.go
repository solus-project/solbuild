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

package main

import (
	_ "builder"
	log "github.com/Sirupsen/logrus"
	"os"
	"solbuild/cmd"
)

// Set up the main logger formatting used in USpin
func init() {
	form := &log.TextFormatter{}
	form.FullTimestamp = true
	form.TimestampFormat = "15:04:05"
	log.SetFormatter(form)
}

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
