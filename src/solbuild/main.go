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
	"flag"
	"fmt"
	"os"
)

var commands map[string]*flag.FlagSet

func init() {
	commands = make(map[string]*flag.FlagSet)
	commands["init"] = flag.NewFlagSet("build", flag.ExitOnError)
	commands["build"] = flag.NewFlagSet("build", flag.ExitOnError)
	commands["update"] = flag.NewFlagSet("update", flag.ExitOnError)
}

func printMainUsage() {
	fmt.Fprintf(os.Stderr, "usage: %v [-h] [command]\n", os.Args[0])
	// TODO: Print command usage!
}

func main() {
	if len(os.Args) < 2 {
		printMainUsage()
	}
}
