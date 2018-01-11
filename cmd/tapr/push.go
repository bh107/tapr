package main

import (
	"flag"
	"io"
	"log"
	"os"

	"hpt.space/tapr"
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
