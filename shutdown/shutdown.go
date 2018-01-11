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

// Package shutdown provides a mechanism for registering handlers to be called
// on process shutdown.
package shutdown // import "hpt.space/tapr/shutdown"

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"hpt.space/tapr/log"
)

// GracePeriod specifies the maximum amount of time during which all shutdown
// handlers must complete before the process forcibly exits.
const GracePeriod = 1 * time.Minute

// Handle registers the onShutdown function to be run when the system is being
// shut down. On shutdown, registered functions are run in last-in-first-out
// order. Handle may be called concurrently.
func Handle(onShutdown func()) {
	shutdown.mu.Lock()
	defer shutdown.mu.Unlock()

	shutdown.sequence = append(shutdown.sequence, onShutdown)
}

// Now calls all registered shutdown closures in last-in-first-out order and
// terminates the process with the given status code.
// It only executes once and guarantees termination within GracePeriod.
// Now may be called concurrently.
func Now(code int) {
	shutdown.once.Do(func() {
		log.Debug.Printf("shutdown: status code %d", code)

		// Ensure we terminate after a fixed amount of time.
		go func() {
			time.Sleep(GracePeriod)
			// Don't use log package here; it may have been flushed already.
			fmt.Fprintf(os.Stderr, "shutdown: %v elapsed since shutdown requested; exiting forcefully", GracePeriod)
			os.Exit(1)
		}()

		shutdown.mu.Lock() // No need to ever unlock.
		for i := len(shutdown.sequence) - 1; i >= 0; i-- {
			shutdown.sequence[i]()
		}

		os.Exit(code)
	})
}

var shutdown struct {
	mu       sync.Mutex
	sequence []func()
	once     sync.Once
}

func init() {
	// close the listener when a shutdown event happens.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt)

	go func() {
		sig := <-c
		log.Error.Printf("shutdown: process received signal %v", sig)
		Now(1)
	}()
}
