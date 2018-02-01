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

	// drive implementation
	_ "tapr.space/store/tape/drive/fake"
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
		stg, err := store.Create(name, cfg.Backend, cfg)
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
