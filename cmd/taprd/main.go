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
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"tapr.space/config"
	"tapr.space/flags"
	"tapr.space/rpc/ioserver"
	"tapr.space/sim"
	"tapr.space/store"

	// store implementations
	_ "tapr.space/store/fs/service"
	_ "tapr.space/store/tape/service"

	// inventory implementations
	_ "tapr.space/store/tape/inv/postgres"

	// changer implementations
	_ "tapr.space/store/tape/changer/fake"
	_ "tapr.space/store/tape/changer/mtx"

	// format implementations
	_ "tapr.space/format/ltfs"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flags.Parse(flags.Server)

	fmt.Println("taprd: starting")

	if flags.Simulate {
		fmt.Println("taprd: simulation enabled")
		sim.Enable()
	}

	fmt.Printf("taprd: server configuration file: %s\n", flags.ServerConfigFile)

	f, err := os.Open(flags.ServerConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	srvConfig, err := config.InitServerConfig(f)
	if err != nil {
		log.Fatal(err)
	}

	for name, cfg := range srvConfig.Stores {
		stg, err := store.Create(name, cfg)
		if err != nil {
			log.Fatal(err)
		}

		// io api server
		httpIO := ioserver.New(config.New(), stg)
		http.Handle("/api/v1/"+name+"/io/", httpIO)
	}

	fmt.Println("taprd: server ready")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
