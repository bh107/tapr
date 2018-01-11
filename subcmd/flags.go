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

// Flag helpers.

package subcmd

import (
	"flag"
	"fmt"
	"os"
)

// ParseFlags parses the flags in the command line arguments,
// according to those set in the flag set.
func (s *State) ParseFlags(fs *flag.FlagSet, args []string, help, usage string) {
	helpFlag := fs.Bool("help", false, "print more information about the command")
	usageFn := func() {
		fmt.Fprintf(os.Stderr, "Usage: tapr %s\n", usage)
		if *helpFlag {
			fmt.Fprintln(os.Stderr, help)
		}
		// How many flags?
		n := 0
		fs.VisitAll(func(*flag.Flag) { n++ })
		if n > 0 {
			fmt.Fprintf(os.Stderr, "Flags:\n")
			fs.PrintDefaults()
		}
	}
	fs.Usage = usageFn
	err := fs.Parse(args)
	if err != nil {
		s.Exit(err)
	}
	if *helpFlag {
		fs.Usage()
		os.Exit(2)
	}
}

// IntFlag returns the value of the named integer flag in the flag set.
func IntFlag(fs *flag.FlagSet, name string) int {
	return fs.Lookup(name).Value.(flag.Getter).Get().(int)
}

// BoolFlag returns the value of the named boolean flag in the flag set.
func BoolFlag(fs *flag.FlagSet, name string) bool {
	return fs.Lookup(name).Value.(flag.Getter).Get().(bool)
}

// StringFlag returns the value of the named string flag in the flag set.
func StringFlag(fs *flag.FlagSet, name string) string {
	return fs.Lookup(name).Value.(flag.Getter).Get().(string)
}
