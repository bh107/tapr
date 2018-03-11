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
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"tapr.space/store/tape"
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
