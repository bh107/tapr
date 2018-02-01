package main

import (
	"flag"
	"io"
	"log"
	"os"

	"tapr.space"
)

func (s *State) pull(args ...string) {
	const help = `
The pull command retrieves a file from the server.

Use the -resume flag to resume an interrupted pull. Resume is only usable if
the -out flag is also specified.
`
	fs := flag.NewFlagSet("pull", flag.ExitOnError)
	outFileFlag := fs.String("out", "", "output file (defaults to standard output)")
	resumeFlag := fs.Bool("resume", false, "resume interrupted pull")
	s.ParseFlags(fs, args, help, "pull [-out=outputfile] path")

	if fs.NArg() != 1 {
		usageAndExit(fs)
	}

	path := tapr.PathName(fs.Arg(0))

	wr := os.Stdout
	var offset int64

	if *outFileFlag != "" {
		if *resumeFlag {
			var err error
			wr, err = os.OpenFile(*outFileFlag, os.O_APPEND|os.O_WRONLY, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}

			offset, err = wr.Seek(0, io.SeekEnd)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			var err error
			wr, err = os.Create(*outFileFlag)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if err := s.Client.PullFile(path, wr, offset); err != nil {
		log.Fatal(err)
	}
}
