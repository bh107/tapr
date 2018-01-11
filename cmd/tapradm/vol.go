package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"hpt.space/tapr/store/tape"
)

func (s *State) vol(args ...string) {
	const help = `
The vol command prints a list of known volumes.
`
	fs := flag.NewFlagSet("vol", flag.ExitOnError)
	longFormat := fs.Bool("l", false, "long format")
	s.ParseFlags(fs, args, help, "vol [-l]")

	vols, err := s.Management.Volumes()
	if err != nil {
		log.Fatal(err)
	}

	if *longFormat {
		tw := tabwriter.NewWriter(os.Stdout, 2, 1, 2, ' ', 0)
		fmt.Fprintf(tw, "SERIAL\tSLOT\tADDR\tHOME\tCATEGORY\tFLAGS\n")
		for _, vol := range vols {
			var home string
			if vol.Home.Category != tape.UnknownSlot {
				home = fmt.Sprintf("%d", vol.Home.Addr)
			}

			fmt.Fprintf(tw, "%v\t%v\t%v\t%s\t%v\t%v\n", vol.Serial, vol.Location.Category, vol.Location.Addr, home, vol.Category, tape.FormatVolumeFlags(vol.Flags))
		}
		tw.Flush()

		return
	}

	for _, vol := range vols {
		fmt.Println(vol.Serial)
	}
}
