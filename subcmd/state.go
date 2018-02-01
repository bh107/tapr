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

package subcmd // import "tapr.space/subcmd"

import (
	"fmt"
	"os"

	"tapr.space"
	"tapr.space/client"
	"tapr.space/mgnt"
	"tapr.space/shutdown"
)

// State describes the state of a subcommand.
// See the comments for Exitf to see how Interactive is used.
// It allows a program to run multiple commands.
type State struct {
	Name       string
	Config     tapr.Config
	Client     tapr.Client
	Management mgnt.Client
	ExitCode   int // Exit with non-zero status for minor problems.
}

// NewState returns a new State for the named subcommand.
func NewState(name string) *State {
	s := &State{Name: name}
	return s
}

// Init initializes the config and client for the State.
func (s *State) Init(config tapr.Config) {
	var c tapr.Client
	var m mgnt.Client
	if config != nil {
		c = client.New(config)
		m = client.NewManagementClient(config)
	}

	s.Config = config
	s.Management = m
	s.Client = c
}

// Exitf prints the error and exits the program.
func (s *State) Exitf(format string, args ...interface{}) {
	format = fmt.Sprintf("tapr: %s: %s\n", s.Name, format)
	fmt.Fprintf(os.Stderr, format, args...)

	s.ExitCode = 1
	s.ExitNow()
}

// Exit calls s.Exitf with the error.
func (s *State) Exit(err error) {
	s.Exitf("%s", err)
}

// ExitNow terminates the process with the current ExitCode.
func (s *State) ExitNow() {
	shutdown.Now(s.ExitCode)
}

// Failf logs the error and sets the exit code. It does not exit the program.
func (s *State) Failf(format string, args ...interface{}) {
	format = fmt.Sprintf("tapr: %s: %s\n", s.Name, format)
	fmt.Fprintf(os.Stderr, format, args...)
	s.ExitCode = 1
}

// Fail calls s.Failf with the error.
func (s *State) Fail(err error) {
	s.Failf("%v", err)
}
