// Copyright (c) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package framework

import (
	"flag"
	"math/rand"
	"time"
)

// Global framework.
var Global *Framework

// Framework handles communication with the kube cluster in e2e tests.
type Framework struct {
	RunID string
}

// Setup sets up a test framework and initialises framework.Global.
func Setup() error {
	// TBD
	runid := flag.String("runid", "", "Optional string that will be used to uniquely identify this test run.")
	flag.Parse()

	if *runid == "" {
		runIDString := "test-" + generateRandomID(3)
		runid = &runIDString
	}

	Global = &Framework{
		RunID: *runid,
	}

	return nil
}

// Teardown shuts down the test framework and cleans up.
func Teardown() error {
	Global = nil
	return nil
}

func generateRandomID(n int) string {
	rand.Seed(time.Now().Unix())
	var letter = []rune("abcdefghijklmnopqrstuvwxyz")

	id := make([]rune, n)
	for i := range id {
		id[i] = letter[rand.Intn(len(letter))]
	}
	return string(id)
}
