// Copyright 2016 The Upspin Authors. All rights reserved.
// Copyright 2017 The Tapr Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
//    * Redistributions of source code must retain the above copyright
//      notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
//      copyright notice, this list of conditions and the following
//      disclaimer in the documentation and/or other materials provided
//      with the distribution.
//    * Neither the name of Google Inc. nor the names of its
//      contributors may be used to endorse or promote products derived
//      from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"tapr.space/config"
	"tapr.space/flags"
	"tapr.space/subcmd"
)

const intro = `
The tapradm command is used for Tapr management.

Each subcommand has a -help flag that explains it in more detail.
For instance

	tapradm vol -help

explains the purpose and usage of the vol subcommand.

There is a set of global flags such as -config to identify the
configuration file to use (default $HOME/.config/tapr/rc) and -log
to set the logging level for debugging. These flags apply across
the subcommands.

Each subcommand has its own set of flags, which if used must appear
after the subcommand name. For example, to run the vol command with
its -l flag and debugging enabled, run

	tapradm -log debug vol -l

For a list of available subcommands and global flags, run

	tapradm -help
`

var commands = map[string]func(*State, ...string){
	"vol": (*State).vol,
}

// State is the command state
type State struct {
	*subcmd.State
	configFile []byte // The contents of the config file we loaded.
}

func main() {
	state, args, ok := setup(flag.CommandLine, os.Args[1:])
	if !ok || len(args) == 0 {
		help()
	}
	if args[0] == "help" {
		help(args[1:]...)
	}

	state.run(args)

	state.ExitNow()
}

// setup initializes the tapr command given the full command-line argument
// list, args. It applies any global flags set on the command line and returns
// the initialized State and the arg list after the global flags, starting with
// the subcommand ("get", "put", etc.) that will be run.
func setup(fs *flag.FlagSet, args []string) (*State, []string, bool) {
	log.SetFlags(0)
	log.SetPrefix("tapradm: ")
	fs.Usage = usage
	flags.ParseArgsInto(fs, args, flags.Client)

	if len(fs.Args()) < 1 {
		return nil, nil, false
	}

	state := newState(strings.ToLower(fs.Arg(0)))
	state.init()

	return state, fs.Args(), true
}

// run runs a single command specified by the arguments, which should begin with
// the subcommand ("ls", "info", etc.).
func (state *State) run(args []string) {
	cmd := state.getCommand(args[0])
	cmd(state, args[1:]...)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of tapradm:\n")
	fmt.Fprintf(os.Stderr, "\ttapradm [globalflags] <command> [flags] <path>\n")
	fmt.Fprintln(os.Stderr)
	printCommands()
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Global flags:\n")
	flag.PrintDefaults()
}

// usageAndExit prints usage message from provided FlagSet,
// and exits the program with status code 2.
func usageAndExit(fs *flag.FlagSet) {
	fs.Usage()
	os.Exit(2)
}

// help prints the help for the arguments provided, or if there is none,
// for the command itself.
func help(args ...string) {
	// Find the first non-flag argument.
	cmd := ""
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			cmd = arg
			break
		}
	}

	if cmd == "" {
		fmt.Fprintln(os.Stderr, intro)
	} else {
		// Simplest solution is re-execing.
		command := exec.Command("tapradm", cmd, "-help")
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		command.Run()
	}

	os.Exit(2)
}

func printCommands() {
	fmt.Fprintf(os.Stderr, "tapradm commands:\n")
	var cmdStrs []string
	for cmd := range commands {
		cmdStrs = append(cmdStrs, cmd)
	}

	sort.Strings(cmdStrs)

	for _, cmd := range cmdStrs {
		fmt.Fprintf(os.Stderr, "\t%s\n", cmd)
	}
}

// getCommand looks up the command named by op.
// If the command can't be found, it exits after listing the commands
// that do exist.
func (state *State) getCommand(op string) func(*State, ...string) {
	op = strings.ToLower(op)
	fn := commands[op]
	if fn != nil {
		return fn
	}

	printCommands()

	state.Exitf("no such command %q", op)

	return nil
}

// newState returns a State with enough initialized to run exit, etc.
// It does not contain a Config.
func newState(name string) *State {
	s := &State{
		State: subcmd.NewState(name),
	}
	return s
}

// init initializes the State with what is required to run the subcommand,
// usually including setting up a Config.
func (state *State) init() {
	data, err := ioutil.ReadFile(flags.Config)
	if err != nil {
		state.Exit(err)
	}

	cfg, err := config.InitConfig(bytes.NewReader(data))

	state.State.Init(cfg)

	state.configFile = data
}
