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
package bitmask_test

import (
	"testing"

	"tapr.space/bitmask"
)

const (
	TestFlagOne uint32 = 1 << iota
	TestFlagTwo
	TestFlagThree
)

func TestAddFlag(t *testing.T) {
	m := TestFlagTwo

	bitmask.Set(&m, TestFlagThree)

	if m != 6 {
		t.Error("flag not set")
	}
}

func TestClearFlag(t *testing.T) {
	m := TestFlagOne | TestFlagThree

	bitmask.Clear(&m, TestFlagOne)

	if m != 4 {
		t.Error("flag not cleared")
	}
}

func TestToggleFlag(t *testing.T) {
	m := TestFlagOne | TestFlagThree

	bitmask.Toggle(&m, TestFlagTwo)

	if m != 7 {
		t.Error("flag not toggled")
	}

	bitmask.Toggle(&m, TestFlagOne)

	if m != 6 {
		t.Error("flag not toggled")
	}
}

func TestHasFlag(t *testing.T) {
	m := TestFlagOne | TestFlagThree

	if bitmask.IsSet(m, TestFlagTwo) {
		t.Error("flag should not have been set")
	}

	bitmask.Set(&m, TestFlagTwo)

	if !bitmask.IsSet(m, TestFlagTwo) {
		t.Error("flag should have been set")
	}
}
