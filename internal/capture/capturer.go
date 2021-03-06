/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package capture

import (
	"github.com/tdimitrov/tranqap/internal/output"
)

const (
	// CapturerDead means that the Capturer has stopped unexpectedly
	CapturerDead = iota
	// CapturerStopped means the Capturer has been stopped by command
	CapturerStopped = iota
)

// CapturerEvent represents the structure of the event generated from Capturer
// to Storage. It has got two parameters:
// from - the address of the Capturer struct in memory. It is used to identify the Capturer
// event - the type of the event. This value should be equal on one of the consts above.
type CapturerEvent struct {
	from  string
	event int
}

// CapturerEventChan is the type of the channel used by MultiOutput for event handling
type CapturerEventChan chan CapturerEvent

// Capturer interface represents a general capturer. There are concrete implementations
// for tcpdump. In the future more can be added, e.g. tshark, dumpcap, etc.
type Capturer interface {
	Start() error
	Stop() error
	AddOutputer(newOutputer output.OutputerFactory) error
	Name() string
}
