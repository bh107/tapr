package ioserver // import "tapr.space/rpc/ioserver"

import (
	"fmt"
	"net/http"
	"sync"

	"tapr.space"
	"tapr.space/log"
	"tapr.space/rpc"
	"tapr.space/store"
)

type server struct {
	config tapr.Config

	st store.Store

	mu struct {
		sync.Mutex

		// open files
		fds map[rpc.Tx]tapr.File
	}
}

// New returns a new http.Handler that presents a storage server
// as a service.
func New(cfg tapr.Config, st store.Store) http.Handler {
	s := &server{
		config: cfg,
		st:     st,
	}

	s.mu.fds = make(map[rpc.Tx]tapr.File)

	return rpc.NewServer(cfg, rpc.Service{
		Name: st.String() + "/io",

		// one-shot methods
		Methods: map[string]rpc.Method{
			"pull/prepare": s.PullPrepare,
			"push/prepare": s.PushPrepare,
		},

		// ingress-based (stream in) methods
		Ingress: map[string]rpc.Ingress{
			"push": s.Push,
		},

		// egress-based (stream out) methods
		Egress: map[string]rpc.Egress{
			"pull":     s.Pull,
			"push/log": s.PushLog,
		},
	})
}

func logf(format string, args ...interface{}) operation {
	s := fmt.Sprintf(format, args...)
	log.Debug.Print("rpc/ioserver: " + s)
	return operation(s)
}

type operation string

func (op operation) log(err error) {
	logf("%v failed: %v", op, err)
}
