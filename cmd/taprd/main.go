package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"hpt.space/tapr/config"
	"hpt.space/tapr/flags"
	"hpt.space/tapr/rpc/ioserver"
	"hpt.space/tapr/sim"
	"hpt.space/tapr/store"

	// store implementations
	_ "hpt.space/tapr/store/fs/service"
	_ "hpt.space/tapr/store/tape/service"

	// inventory implementations
	_ "hpt.space/tapr/store/tape/inv/postgres"

	// changer implementations
	_ "hpt.space/tapr/store/tape/changer/fake"
	_ "hpt.space/tapr/store/tape/changer/mtx"
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
