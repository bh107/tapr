// Copyright 2018 Klaus Birkelund Abildgaard Jensen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"io"
	"log"
	"os"

	"tapr.space"
)

func (s *State) push(args ...string) {
	const help = `
The push command stores a file on the server.

To resume a failed push, add the -resume flag. To append to a previously
stored file add the -append flag. Note that -resume and -append are mutually
exclusive.
`
	fs := flag.NewFlagSet("push", flag.ExitOnError)
	inFileFlag := fs.String("in", "", "input file (defaults to standard input)")
	appendFlag := fs.Bool("append", false, "append data")
	resumeFlag := fs.Bool("resume", false, "resume interrupted push")
	s.ParseFlags(fs, args, help, "push [-in=inputfile] [-append] [-resume] name")

	if fs.NArg() != 1 {
		usageAndExit(fs)
	}

	if *appendFlag && *resumeFlag {
		log.Fatal("error: cannot use both -append and -resume")
	}

	name := tapr.PathName(fs.Arg(0))

	rd := os.Stdin

	if *inFileFlag != "" {
		var err error
		rd, err = os.Open(*inFileFlag)
		if err != nil {
			log.Fatal(err)
		}
	}

	// If resume is given, perform a stat to get the size of the stored file.
	// Use the information to advance the reader.
	if *resumeFlag {
		fileInfo, err := s.Client.Stat(name)
		if err != nil {
			log.Fatal(err)
		}

		// seek locally
		_, err = rd.Seek(fileInfo.Size, io.SeekStart)
		if err != nil {
			log.Fatal(err)
		}

		// now just perform an append
		if err := s.Client.Append(name, rd); err != nil {
			log.Fatal(err)
		}

		return
	}

	if err := s.Client.PushFile(name, rd, *appendFlag); err != nil {
		log.Fatal(err)
	}
}
