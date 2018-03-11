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

package sim

import (
	"math/rand"
	"sync/atomic"
	"time"

	"tapr.space/log"
)

var enabled int32 // read/written atomically

// Enable enables simulation.
func Enable() {
	if !atomic.CompareAndSwapInt32(&enabled, 0, 1) {
		panic("already enabled")
	}
}

// State is a simulation state.
type State interface {
	// Simulate uses the given Noise to simulate some duration.
	Simulate(n Noise)
}

// noop is a simple no-op implementation of the simulation state.
type noop struct{}

func (noop) Simulate(Noise) {}

type state struct {
	shutdown chan struct{}
}

var defaultState = &state{
	shutdown: make(chan struct{}),
}

// Noise represents a random distribution.
type Noise interface {
	// Generate generates a time.Duration related to the
	// noise distribution.
	Generate() time.Duration
}

// NormalDistributedNoise is a struct. Doh.
type NormalDistributedNoise struct {
	Mean, Stddev time.Duration
}

// Generate implements the Noiser interface.
func (n *NormalDistributedNoise) Generate() time.Duration {
	return n.Mean + time.Duration(rand.NormFloat64()*float64(n.Stddev))
}

// Simulate generates noise from the given Noise and sleeps for that
// duration.
func (s *state) Simulate(n Noise) {
	select {
	case <-time.After(n.Generate()):
		return

	case <-s.shutdown:
		log.Info.Print("sim: shutting down; stopping simulations")
		return
	}
}

// Enabled returns whether simulation is enabled.
func Enabled() bool {
	return atomic.LoadInt32(&enabled) == 1
}

// Maybe executes the given fn in a simulation state.
func Maybe(fn func(s State)) {
	// check if simulation is enabled
	if atomic.LoadInt32(&enabled) == 0 {
		// execute in normal state
		fn(&noop{})
		return
	}

	// execute in simulated state
	fn(defaultState)
}

// Maybe executes the given fn in a simulation state.
func Maybex(s State, fn func(State)) {
	// check if simulation is enabled
	if atomic.LoadInt32(&enabled) == 0 {
		// execute in normal state
		fn(&noop{})
		return
	}

	// execute in simulated state
	fn(s)
}
