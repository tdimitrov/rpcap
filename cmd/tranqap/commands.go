/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package main

import (
	"github.com/abiosoft/ishell"
	"github.com/tdimitrov/tranqap/internal/capture"
	"github.com/tdimitrov/tranqap/internal/output"
	"github.com/tdimitrov/tranqap/internal/tqlog"
)

const (
	cmdErr  = iota
	cmdOk   = iota
	cmdExit = iota
)

var capturers *capture.Storage

func initStorage() {
	capturers = capture.NewStorage()
}

func getSudoConfig(t target) capture.SudoConfig {
	var ret capture.SudoConfig
	if *t.UseSudo == true {
		ret.Use = true
		ret.Username = new(string)
		*ret.Username = *t.User
	} else {
		ret.Use = false
		ret.Username = nil
	}

	return ret
}

func getFilterConfig(t target) capture.FilterConfig {
	return capture.FilterConfig{Port: t.FilterPort}
}

func cmdStart(ctx *ishell.Context, cfg configParams) {
	// Check if there is a running job
	if capturers.Empty() == false {
		ctx.Println("There is already a running capture")
		return
	}

	tqlog.Info("Called start command")

	for _, t := range cfg.Targets {
		c, d, err := getClientConfig(&t)
		if err != nil {
			ctx.Printf("Error parsing client configuration for target <%s>: %s\n", *t.Name, err)
			return
		}

		// Create file output
		f := output.NewFileOutput(*t.Destination, *t.FilePattern, *t.RotationCnt)
		if f == nil {
			ctx.Printf("Can't create File output for target <%s>\n", *t.Name)
			return
		}

		// Create multioutput and attach the file output to it
		m := output.NewMultiOutput(f)
		if m == nil {
			ctx.Printf("Can't create MultiOutput for target <%s>\n.", *t.Name)
			return
		}

		// Create SSH client
		sshClient := NewSSHClient(*d, *c)

		// Create capturer
		capt := capture.NewTcpdump(*t.Name, m, capturers.GetChan(), sshClient, getSudoConfig(t), getFilterConfig(t))
		if capt == nil {
			ctx.Printf("Error creating Capturer for target <%s>\n", *t.Name)
			return
		}

		if err := capt.Start(); err != nil {
			ctx.Println(err)
			return
		}

		if err := capturers.Add(capt); err != nil {
			ctx.Printf("Error adding capturer: %s", err.Error())
		}
	}
}

func cmdStop(ctx *ishell.Context) {
	// Check if there is a running job
	if capturers.Empty() == true {
		ctx.Println("There are no running captures.")
		return
	}

	tqlog.Info("Called stop command")

	capturers.StopAll()
}

func cmdWireshark(ctx *ishell.Context) {
	tqlog.Info("Called wireshark command with args %v", ctx.Args)

	// Prepare a factory function, which creates Wireshark Outputer
	factFn := func(p output.MOEventChan) output.Outputer {
		return output.NewWsharkOutput(p)
	}

	capturers.AddNewOutput(factFn, ctx.Args)
}

func cmdTargets(ctx *ishell.Context, cfg configParams) {
	tqlog.Info("Called targets command")

	for _, t := range cfg.Targets {
		c, d, err := getClientConfig(&t)
		if err != nil {
			ctx.Printf("Error parsing client configuration for target <%s>: %s\n", *t.Name, err)
			return
		}

		ctx.Printf("=== Running checks for target <%s> ===\n", *t.Name)
		sshClient := NewSSHClient(*d, *c)
		if output, err := checkPermissions(sshClient); err != nil {
			ctx.Printf("%s\n", err)
		} else {
			ctx.Printf("%s\n", output)
		}
	}

	return
}
