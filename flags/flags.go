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

// Package flags defines command-line flags to make them consistent between
// binaries. Not all flags make sense for all binaries.
package flags // import "tapr.space/flags"

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tapr.space/config"
	"tapr.space/log"
)

// flagVar represents a flag in this package.
type flagVar struct {
	set func(fs *flag.FlagSet) // Set the value at parse time.
	arg func() string          // Return the argument to set the flag.
}

const (
	defaultHTTPAddr = ":8080"
	defaultLog      = "info"
)

var (
	defaultDataDir          = taprDir("data")
	defaultConfigFile       = taprDir("config")
	defaultServerConfigFile = taprDir("server/config")
)

func taprDir(target string) string {
	home, err := config.Homedir()
	if err != nil {
		log.Error.Printf("flags: could not locate home directory: %v", err)
		home = "."
	}

	return filepath.Join(home, "tapr", target)
}

// None is the set of no flags. It is rarely needed as most programs
// use either the Server or Client set.
var None = []string{}

// Server is the set of flags most useful in servers. It can be passed as the
// argument to Parse to set up the package for a server.
var Server = []string{
	"log", "simulate", "emulate-dev", "dbreset", "audit", "serverconfig",
}

// Client is the set of flags most useful in clients. It can be passed as the
// argument to Parse to set up the package for a client.
var Client = []string{
	"config", "log", "store",
}

var (
	// Config ("config") names the configuration file to use.
	Config = defaultConfigFile

	// ServerConfigFile is the server configuration file
	ServerConfigFile = defaultServerConfigFile

	// HTTPAddr ("http") is the network address on which to listen for
	// incoming insecure network connections.
	HTTPAddr = defaultHTTPAddr

	// Store is the store to target.
	Store = "default"

	// Log ("log") sets the level of logging (implements flag.Value).
	Log logFlag

	// Audit ("audit") request that an audit of the inventory is performed
	// when initializing.
	Audit = false

	// ResetDB ("dbreset") resets the inventory database
	ResetDB = false

	// Simulate ("simulate") enable simulation
	Simulate = false

	// EmulateDevices ("emulate-dev") enable emulation of devices
	EmulateDevices = false

	// Version causes the program to print its release version and exit.
	// The printed version is only meaningful in released binaries.
	Version = false
)

// flags is a map of flag registration functions keyed by flag name,
// used by Parse to register specific (or all) flags.
var flags = map[string]*flagVar{
	"config": strVar(&Config, "config", Config, "configuration `file`"),
	"http":   strVar(&HTTPAddr, "http", HTTPAddr, "`address` for incoming insecure network connections"),
	"store":  strVar(&Store, "store", Store, "store to target"),

	"log": {
		set: func(fs *flag.FlagSet) {
			Log.Set("info")
			fs.Var(&Log, "log", "`level` of logging: debug, info, warning, error, disabled")
		},
		arg: func() string { return strArg("log", Log.String(), defaultLog) },
	},

	"audit":   boolVar(&Audit, "audit", false, "whether to perform in inventory audit"),
	"dbreset": boolVar(&ResetDB, "dbreset", false, "whether to reset the inventory database"),

	"serverconfig": strVar(&ServerConfigFile, "serverconfig", ServerConfigFile, "server configuration `file`"),

	"simulate":    boolVar(&Simulate, "simulate", false, "whether to enable simulation of operations"),
	"emulate-dev": boolVar(&EmulateDevices, "emulate-dev", false, "whether to enable emulation of devices"),
	"version":     boolVar(&Version, "version", false, "print build version and exit"),
}

// Parse registers the command-line flags for the given default flags list, plus
// any extra flag names, and calls flag.Parse. Passing no flag names in either
// list registers all flags. Passing an unknown name triggers a panic.
// The Server and Client variables contain useful default sets.
//
// Examples:
//      flags.Parse(flags.Client) // Register all client flags.
//      flags.Parse(flags.Server, "cachedir") // Register all server flags plus cachedir.
//      flags.Parse(nil) // Register all flags.
//      flags.Parse(flags.None, "config", "endpoint") // Register only config and endpoint.
func Parse(defaultList []string, extras ...string) {
	ParseArgsInto(flag.CommandLine, os.Args[1:], defaultList, extras...)
}

// ParseInto is the same as Parse but accepts a FlagSet argument instead of
// using the default flag.CommandLine FlagSet.
func ParseInto(fs *flag.FlagSet, defaultList []string, extras ...string) {
	ParseArgsInto(fs, os.Args[1:], defaultList, extras...)
}

// ParseArgs is the same as Parse but uses the provided argument list
// instead of those provided on the command line. For ParseArgs, the
// initial command name should not be provided.
func ParseArgs(args, defaultList []string, extras ...string) {
	ParseArgsInto(flag.CommandLine, args, defaultList, extras...)
}

// ParseArgsInto is the same as ParseArgs but accepts a FlagSet argument instead of
// using the default flag.CommandLine FlagSet.
func ParseArgsInto(fs *flag.FlagSet, args, defaultList []string, extras ...string) {
	if len(defaultList) == 0 && len(extras) == 0 {
		RegisterInto(fs)
	} else {
		if len(defaultList) > 0 {
			RegisterInto(fs, defaultList...)
		}
		if len(extras) > 0 {
			RegisterInto(fs, extras...)
		}
	}

	fs.Parse(args)
}

// Register registers the command-line flags for the given flag names.
// Unlike Parse, it may be called multiple times.
// Passing zero names install all flags.
// Passing an unknown name triggers a panic.
//
// For example:
//      flags.Register("config", "endpoint") // Register Config and Endpoint.
// or
//      flags.Register() // Register all flags.
func Register(names ...string) {
	RegisterInto(flag.CommandLine, names...)
}

// RegisterInto  is the same as Register but accepts a FlagSet argument instead of
// using the default flag.CommandLine FlagSet.
func RegisterInto(fs *flag.FlagSet, names ...string) {
	if len(names) == 0 {
		// Register all flags if no names provided.
		for _, f := range flags {
			f.set(fs)
		}
	} else {
		for _, n := range names {
			f, ok := flags[n]
			if !ok {
				panic(fmt.Sprintf("unknown flag %q", n))
			}
			f.set(fs)
		}
	}
}

// Args returns a slice of -flag=value strings that will recreate
// the state of the flags. Flags set to their default value are elided.
func Args() []string {
	var args []string
	for _, f := range flags {
		arg := f.arg()
		if arg == "" {
			continue
		}
		args = append(args, arg)
	}

	return args
}

// strVar returns a flagVar for the given string flag.
func strVar(value *string, name, _default, usage string) *flagVar {
	return &flagVar{
		set: func(fs *flag.FlagSet) {
			fs.StringVar(value, name, _default, usage)
		},
		arg: func() string {
			return strArg(name, *value, _default)
		},
	}
}

// strArg returns a command-line argument that will recreate the flag,
// or the empty string if the value is the default.
func strArg(name, value, _default string) string {
	if value == _default {
		return ""
	}
	return "-" + name + "=" + value
}

// boolVar returns a flagVar for the given boolean flag.
func boolVar(value *bool, name string, _default bool, usage string) *flagVar {
	return &flagVar{
		set: func(fs *flag.FlagSet) {
			fs.BoolVar(value, name, _default, usage)
		},
		arg: func() string {
			return boolArg(name, *value, _default)
		},
	}
}

// boolArg returns a command-line argument that will recreate the flag,
// or the empty string if the value is the default.
func boolArg(name string, value, _default bool) string {
	if value == _default {
		return ""
	}
	return "-" + name
}

type logFlag string

// String implements flag.Value.
func (f logFlag) String() string {
	return string(f)
}

// Set implements flag.Value.
func (f *logFlag) Set(level string) error {
	err := log.SetLevel(level)
	if err != nil {
		return err
	}

	*f = logFlag(log.GetLevel())

	return nil
}

// Get implements flag.Getter.
func (logFlag) Get() interface{} {
	return log.GetLevel()
}

type configFlag struct {
	s *[]string
}

// String implements flag.Value.
func (f configFlag) String() string {
	if f.s == nil {
		return ""
	}
	return strings.Join(*f.s, ",")
}

// Set implements flag.Value.
func (f configFlag) Set(s string) error {
	ss := strings.Split(strings.TrimSpace(s), ",")
	// Drop empty elements.
	for i := 0; i < len(ss); i++ {
		if ss[i] == "" {
			ss = append(ss[:i], ss[i+1:]...)
		}
	}
	*f.s = ss
	return nil
}

// Get implements flag.Getter.
func (f configFlag) Get() interface{} {
	if f.s == nil {
		return ""
	}
	return *f.s
}
