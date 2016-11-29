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
	flag "github.com/spf13/pflag"
	"os"
)

var commands map[string]*flag.FlagSet
var profile string

func addProfileFlag(f *flag.FlagSet) {
	f.StringVarP(&profile, "profile", "p", "main", "build profile to use")
}

func init() {
	commands = make(map[string]*flag.FlagSet)
	commands["init"] = flag.NewFlagSet("build", flag.ExitOnError)
	commands["build"] = flag.NewFlagSet("build", flag.ExitOnError)
	commands["update"] = flag.NewFlagSet("update", flag.ExitOnError)

	addProfileFlag(commands["init"])
	addProfileFlag(commands["build"])
	addProfileFlag(commands["update"])
}

func printMainUsage() {
	fmt.Fprintf(os.Stderr, "usage: %v [-h] [command]\n", os.Args[0])
	// TODO: Print command usage!
}

func main() {
	if len(os.Args) < 2 {
		printMainUsage()
		os.Exit(1)
	}
	arg := os.Args[1]
	cmd, found := commands[arg]
	if !found {
		fmt.Fprintf(os.Stderr, "Unknown command: %v\n", arg)
		printMainUsage()
		os.Exit(1)
	}
	// Parse beyond subcommand
	cmd.Parse(os.Args[2:])

	fmt.Fprintf(os.Stderr, "Selected %v\n", profile)
}
